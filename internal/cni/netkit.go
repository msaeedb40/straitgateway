package cni

import (
	"crypto/rand"
	"fmt"
	"net"
	"runtime"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// DeviceType defines the interface driver used
type DeviceType string

const (
	DeviceTypeNetkit DeviceType = "netkit"
	DeviceTypeVeth   DeviceType = "veth"
)

// PodInterfaceConfig defines parameters for configuring a pod network interface
type PodInterfaceConfig struct {
	HostIfName      string
	ContainerIfName string
	NetNSPath       string
	IPv4            net.IPNet
	IPv6            *net.IPNet
	MTU             int
	GatewayV4       net.IP
	GatewayV6       net.IP
}

// InterfaceResult holds metadata for configured interface pair
type InterfaceResult struct {
	DeviceType       DeviceType
	HostInterface    string
	HostIfIndex      int
	HostMAC          string
	ContainerIfIndex int
	ContainerMAC     string
}

// SetupPodInterface creates the host/container interface pair and configures pod networking
func SetupPodInterface(cfg PodInterfaceConfig) (*InterfaceResult, error) {
	if cfg.MTU == 0 {
		cfg.MTU = 1500
	}
	if cfg.ContainerIfName == "" {
		cfg.ContainerIfName = "eth0"
	}
	if cfg.HostIfName == "" {
		cfg.HostIfName = generateHostIfName()
	}

	// Try creating netkit pair first; if not supported in kernel, fallback to veth
	devType := DeviceTypeNetkit
	err := createNetkitPair(cfg.HostIfName, cfg.ContainerIfName)
	if err != nil {
		devType = DeviceTypeVeth
		if errVeth := createVethPair(cfg.HostIfName, cfg.ContainerIfName); errVeth != nil {
			return nil, fmt.Errorf("failed to create netkit (%v) and veth fallback (%v)", err, errVeth)
		}
	}

	// Retrieve host link
	hostLink, err := netlink.LinkByName(cfg.HostIfName)
	if err != nil {
		return nil, fmt.Errorf("find host link %s: %w", cfg.HostIfName, err)
	}

	// Retrieve peer link before moving to container namespace
	peerLink, err := netlink.LinkByName(cfg.ContainerIfName)
	if err != nil {
		_ = netlink.LinkDel(hostLink)
		return nil, fmt.Errorf("find peer link %s: %w", cfg.ContainerIfName, err)
	}

	// Open container netns
	targetNs, err := netns.GetFromPath(cfg.NetNSPath)
	if err != nil {
		_ = netlink.LinkDel(hostLink)
		return nil, fmt.Errorf("open netns %s: %w", cfg.NetNSPath, err)
	}
	defer targetNs.Close()

	// Move peer interface to container netns
	if err := netlink.LinkSetNsFd(peerLink, int(targetNs)); err != nil {
		_ = netlink.LinkDel(hostLink)
		return nil, fmt.Errorf("move link to netns: %w", err)
	}

	// Bring up host link
	_ = netlink.LinkSetMTU(hostLink, cfg.MTU)
	if err := netlink.LinkSetUp(hostLink); err != nil {
		_ = netlink.LinkDel(hostLink)
		return nil, fmt.Errorf("bring up host link: %w", err)
	}

	// Enter container netns to configure address and routes
	var containerIfIndex int
	var containerMAC string

	err = executeInNetNS(targetNs, func() error {
		link, err := netlink.LinkByName(cfg.ContainerIfName)
		if err != nil {
			return fmt.Errorf("find link in netns: %w", err)
		}
		containerIfIndex = link.Attrs().Index
		containerMAC = link.Attrs().HardwareAddr.String()

		_ = netlink.LinkSetMTU(link, cfg.MTU)

		// Assign IPv4
		addrV4 := &netlink.Addr{IPNet: &cfg.IPv4}
		if err := netlink.AddrAdd(link, addrV4); err != nil {
			return fmt.Errorf("add IPv4 address: %w", err)
		}

		// Assign IPv6 if provided
		if cfg.IPv6 != nil {
			addrV6 := &netlink.Addr{IPNet: cfg.IPv6}
			if err := netlink.AddrAdd(link, addrV6); err != nil {
				return fmt.Errorf("add IPv6 address: %w", err)
			}
		}

		// Bring up container interface
		if err := netlink.LinkSetUp(link); err != nil {
			return fmt.Errorf("bring up container interface: %w", err)
		}

		// Bring up loopback
		lo, err := netlink.LinkByName("lo")
		if err == nil {
			_ = netlink.LinkSetUp(lo)
		}

		// Add default IPv4 route if gateway is present
		if cfg.GatewayV4 != nil {
			gwRoute := &netlink.Route{
				LinkIndex: link.Attrs().Index,
				Gw:        cfg.GatewayV4,
			}
			_ = netlink.RouteAdd(gwRoute)
		}
		return nil
	})

	if err != nil {
		_ = netlink.LinkDel(hostLink)
		return nil, fmt.Errorf("configure container networking: %w", err)
	}

	return &InterfaceResult{
		DeviceType:       devType,
		HostInterface:    cfg.HostIfName,
		HostIfIndex:      hostLink.Attrs().Index,
		HostMAC:          hostLink.Attrs().HardwareAddr.String(),
		ContainerIfIndex: containerIfIndex,
		ContainerMAC:     containerMAC,
	}, nil
}

// TeardownPodInterface deletes the host end of the netkit/veth pair
func TeardownPodInterface(hostIfName string) error {
	link, err := netlink.LinkByName(hostIfName)
	if err != nil {
		return nil // Already gone
	}
	return netlink.LinkDel(link)
}

func createNetkitPair(hostName, peerName string) error {
	// Use netlink generic link for netkit type
	la := netlink.NewLinkAttrs()
	la.Name = hostName
	netkitLink := &netlink.GenericLink{
		LinkAttrs: la,
		LinkType:  "netkit",
	}
	return netlink.LinkAdd(netkitLink)
}

func createVethPair(hostName, peerName string) error {
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: hostName,
		},
		PeerName: peerName,
	}
	return netlink.LinkAdd(veth)
}

func generateHostIfName() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("sg%x", b)
}

func executeInNetNS(targetNs netns.NsHandle, fn func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origNs, err := netns.Get()
	if err != nil {
		return err
	}
	defer origNs.Close()

	if err := netns.Set(targetNs); err != nil {
		return err
	}
	defer netns.Set(origNs) // restore original namespace

	return fn()
}
