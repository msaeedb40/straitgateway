package cni

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// DeviceType defines the interface driver used
type DeviceType string

const (
	DeviceTypeNetkit DeviceType = "netkit"
	DeviceTypeVeth   DeviceType = "veth"

	DefaultStateDir = "/var/run/straitgateway/cni"
)

// PodInterfaceConfig defines parameters for configuring a pod network interface
type PodInterfaceConfig struct {
	ContainerID     string
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

// PodMetadata stores persistent runtime information for CNI CHECK and DEL operations
type PodMetadata struct {
	ContainerID     string     `json:"containerId"`
	HostIfName      string     `json:"hostIfName"`
	ContainerIfName string     `json:"containerIfName"`
	NetNSPath       string     `json:"netnsPath"`
	IPv4            string     `json:"ipv4"`
	IPv6            string     `json:"ipv6,omitempty"`
	DeviceType      DeviceType `json:"deviceType"`
	CreatedAt       string     `json:"createdAt"`
}

// DeriveHostIfName generates a deterministic, leak-free Linux network interface name from a container ID.
// Max 12 characters ("sg" + 10 hex digits), well below the 15-character Linux IFNAMSIZ limit.
func DeriveHostIfName(containerID string) string {
	if containerID == "" {
		return "sg0000000000"
	}
	hash := sha256.Sum256([]byte(containerID))
	return fmt.Sprintf("sg%s", hex.EncodeToString(hash[:5]))
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
		cfg.HostIfName = DeriveHostIfName(cfg.ContainerID)
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
	defer func() { _ = targetNs.Close() }()

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

	// Save pod metadata to disk for CNI CHECK / DEL
	meta := PodMetadata{
		ContainerID:     cfg.ContainerID,
		HostIfName:      cfg.HostIfName,
		ContainerIfName: cfg.ContainerIfName,
		NetNSPath:       cfg.NetNSPath,
		IPv4:            cfg.IPv4.String(),
		DeviceType:      devType,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
	}
	if cfg.IPv6 != nil {
		meta.IPv6 = cfg.IPv6.String()
	}
	_ = savePodMetadata(meta)

	return &InterfaceResult{
		DeviceType:       devType,
		HostInterface:    cfg.HostIfName,
		HostIfIndex:      hostLink.Attrs().Index,
		HostMAC:          hostLink.Attrs().HardwareAddr.String(),
		ContainerIfIndex: containerIfIndex,
		ContainerMAC:     containerMAC,
	}, nil
}

// TeardownPodInterface deletes the host end of the netkit/veth pair and removes saved metadata
func TeardownPodInterface(containerID, hostIfName string) error {
	if hostIfName == "" && containerID != "" {
		hostIfName = DeriveHostIfName(containerID)
	}

	if hostIfName != "" {
		link, err := netlink.LinkByName(hostIfName)
		if err == nil {
			_ = netlink.LinkDel(link)
		}
	}

	if containerID != "" {
		_ = deletePodMetadata(containerID)
	}
	return nil
}

func createNetkitPair(hostName, _ string) error {
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

func savePodMetadata(meta PodMetadata) error {
	if meta.ContainerID == "" {
		return nil
	}
	if err := os.MkdirAll(DefaultStateDir, 0755); err != nil {
		return err
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	filePath := filepath.Join(DefaultStateDir, fmt.Sprintf("%s.json", meta.ContainerID))
	return os.WriteFile(filePath, data, 0644)
}

// GetPodMetadata reads saved container metadata
func GetPodMetadata(containerID string) (*PodMetadata, error) {
	filePath := filepath.Join(DefaultStateDir, fmt.Sprintf("%s.json", containerID))
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var meta PodMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func deletePodMetadata(containerID string) error {
	filePath := filepath.Join(DefaultStateDir, fmt.Sprintf("%s.json", containerID))
	return os.Remove(filePath)
}

func executeInNetNS(targetNs netns.NsHandle, fn func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origNs, err := netns.Get()
	if err != nil {
		return err
	}
	defer func() { _ = origNs.Close() }()

	if err := netns.Set(targetNs); err != nil {
		return err
	}
	defer func() { _ = netns.Set(origNs) }() // restore original namespace

	return fn()
}
