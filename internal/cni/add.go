package cni

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/straitgateway/straitgateway/internal/cni/ipam"
	"github.com/straitgateway/straitgateway/internal/linux/link"
	"github.com/straitgateway/straitgateway/internal/linux/namespace"
	"github.com/straitgateway/straitgateway/internal/tgwd/api"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

// CmdAdd executes the CNI ADD operation with transactional rollback on failure.
func CmdAdd(stdin io.Reader) error {
	var netConf NetConf
	if err := json.NewDecoder(stdin).Decode(&netConf); err != nil {
		return fmt.Errorf("cni: failed to parse netconf from stdin: %w", err)
	}

	env, err := ParseCNIEnv()
	if err != nil {
		return fmt.Errorf("cni: failed to parse environment: %w", err)
	}

	if env.NetNS == "" {
		return fmt.Errorf("cni: CNI_NETNS is required for ADD")
	}
	if env.ContainerID == "" {
		return fmt.Errorf("cni: CNI_CONTAINERID is required for ADD")
	}

	// 1. Setup IPAM (Principal Rule 3: No PodCIDR/RFC1918 assumption)
	cidrs := make([]string, 0)
	for _, r := range netConf.IPAM.Ranges {
		if r.Subnet != "" {
			cidrs = append(cidrs, r.Subnet)
		}
	}
	if len(cidrs) == 0 && env.ParsedArgs.PodCIDR != "" {
		for _, part := range strings.Split(env.ParsedArgs.PodCIDR, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				cidrs = append(cidrs, part)
			}
		}
	}
	if len(cidrs) == 0 {
		return fmt.Errorf("cni: no PodCIDR configured in IPAM ranges or CNI_ARGS (StraitGateway Rule 3 prohibits assuming RFC1918/default CIDRs)")
	}

	store := ipam.NewStore("")
	allocator, err := ipam.NewAllocator(cidrs)
	if err != nil {
		return fmt.Errorf("cni: init ipam allocator: %w", err)
	}
	if err := allocator.RestoreFromStore(store); err != nil {
		zap.L().Warn("cni: failed to restore allocations from store",
			zap.String("containerID", env.ContainerID),
			zap.Error(err),
		)
	}

	// 2. Allocate IP(s)
	allocation, err := allocator.AllocateDualStack()
	if err != nil {
		return fmt.Errorf("cni: ip allocation failed: %w", err)
	}

	// Setup transaction rollback tracking
	rollbackIP := true
	rollbackLink := false
	hostIfName := HostInterfaceName(env.ContainerID)
	tempPeerName := "sgp-" + env.ContainerID[:min(10, len(env.ContainerID))]

	defer func() {
		if rollbackLink {
			_ = link.DeleteNetkit(hostIfName)
		}
		if rollbackIP {
			if allocation.IPv4 != "" {
				_ = allocator.Release(allocation.IPv4)
			}
			if allocation.IPv6 != "" {
				_ = allocator.Release(allocation.IPv6)
			}
		}
	}()

	// 3. Create Netkit link (or veth fallback)
	mtu := netConf.MTU
	if mtu <= 0 {
		mtu = 1500
	}

	nk, err := link.CreateNetkit(hostIfName, tempPeerName, mtu)
	if err != nil {
		return fmt.Errorf("cni: create netkit pair (%s, %s): %w", hostIfName, tempPeerName, err)
	}
	rollbackLink = true

	// Bring host-side interface up
	if err := link.SetLinkUp(hostIfName); err != nil {
		return fmt.Errorf("cni: set host interface %s up: %w", hostIfName, err)
	}

	// 4. Open pod network namespace
	ns, err := namespace.GetFromPath(env.NetNS)
	if err != nil {
		return fmt.Errorf("cni: open pod netns %s: %w", env.NetNS, err)
	}
	defer ns.Close()

	// 5. Move peer interface into container netns
	if err := link.MoveLinkToNetNS(nk.Peer, int(nk.Peer.Attrs().Index)); err != nil {
		// Use netlink LinkSetNsFd
		fdLink, errOpen := os.Open(env.NetNS)
		if errOpen != nil {
			return fmt.Errorf("cni: open netns fd: %w", errOpen)
		}
		defer fdLink.Close()
		if err := netlink.LinkSetNsFd(nk.Peer, int(fdLink.Fd())); err != nil {
			return fmt.Errorf("cni: move peer link into netns: %w", err)
		}
	}

	// 6. Configure network inside pod namespace
	var podMac string
	err = ns.Do(func() error {
		// Set loopback UP
		lo, err := netlink.LinkByName("lo")
		if err == nil {
			_ = netlink.LinkSetUp(lo)
		}

		// Rename temp peer to target container interface name (e.g. eth0)
		peerLink, err := netlink.LinkByName(tempPeerName)
		if err != nil {
			return fmt.Errorf("lookup temp peer %s inside netns: %w", tempPeerName, err)
		}
		if err := netlink.LinkSetName(peerLink, env.IfName); err != nil {
			return fmt.Errorf("rename %s to %s: %w", tempPeerName, env.IfName, err)
		}

		// Re-fetch renamed link
		contLink, err := netlink.LinkByName(env.IfName)
		if err != nil {
			return fmt.Errorf("lookup container link %s: %w", env.IfName, err)
		}
		podMac = contLink.Attrs().HardwareAddr.String()

		// Configure IPv4
		if allocation.IPv4 != "" {
			if err := link.AddAddr(env.IfName, allocation.IPv4); err != nil {
				return fmt.Errorf("add IPv4 %s: %w", allocation.IPv4, err)
			}
		}

		// Configure IPv6
		if allocation.IPv6 != "" {
			if err := link.AddAddr(env.IfName, allocation.IPv6); err != nil {
				return fmt.Errorf("add IPv6 %s: %w", allocation.IPv6, err)
			}
		}

		// Bring interface UP
		if err := netlink.LinkSetUp(contLink); err != nil {
			return fmt.Errorf("set container interface %s up: %w", env.IfName, err)
		}

		// Add default route(s)
		if allocation.Gateway4 != nil {
			if err := link.AddDefaultRoute(env.IfName, allocation.Gateway4); err != nil {
				return fmt.Errorf("add IPv4 default route via %s: %w", allocation.Gateway4, err)
			}
		}
		if allocation.Gateway6 != nil {
			if err := link.AddDefaultRoute(env.IfName, allocation.Gateway6); err != nil {
				return fmt.Errorf("add IPv6 default route via %s: %w", allocation.Gateway6, err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cni: configure container netns: %w", err)
	}

	// 7. Persist allocation
	if err := store.Put(ipam.PodAllocation{
		ContainerID:  env.ContainerID,
		PodName:      env.ParsedArgs.PodName,
		PodNamespace: env.ParsedArgs.PodNamespace,
		IfName:       env.IfName,
		IPv4:         allocation.IPv4,
		IPv6:         allocation.IPv6,
		AllocatedAt:  time.Now().Unix(),
	}); err != nil {
		zap.L().Warn("cni: failed to persist allocation to store",
			zap.String("containerID", env.ContainerID),
			zap.Error(err),
		)
	}

	// 8. Register with StraitD (resilient - non-fatal if StraitD is initializing, but retried and logged)
	socketPath := ResolveSocketPath(netConf.StraitDSocket)
	go func() {
		var lastErr error
		for attempt := 0; attempt < 3; attempt++ {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			client, err := api.NewClient(ctx, zap.NewNop(), socketPath)
			if err != nil {
				lastErr = err
				cancel()
				time.Sleep(500 * time.Millisecond)
				continue
			}

			ips := []string{}
			if allocation.IPv4 != "" {
				ips = append(ips, allocation.IPv4)
			}
			if allocation.IPv6 != "" {
				ips = append(ips, allocation.IPv6)
			}
			_, err = client.ApplyPodNetwork(ctx, &api.ApplyPodNetworkRequest{
				Header: &api.ResourceHeader{
					ResourceUid: env.ContainerID,
					Generation:  1,
				},
				Pod: &api.PodNetwork{
					ContainerId:  env.ContainerID,
					NetnsPath:    env.NetNS,
					PodName:      env.ParsedArgs.PodName,
					PodNamespace: env.ParsedArgs.PodNamespace,
					IfName:       env.IfName,
					IpAddresses:  ips,
					MacAddress:   podMac,
					HostIfIndex:  uint32(nk.Parent.Attrs().Index),
				},
			})
			_ = client.Close()
			cancel()
			if err == nil {
				return
			}
			lastErr = err
			time.Sleep(500 * time.Millisecond)
		}
		zap.L().Warn("cni: failed to register pod with straitd after retries",
			zap.String("containerID", env.ContainerID),
			zap.Error(lastErr),
		)
	}()

	// Success! Disable rollbacks
	rollbackIP = false
	rollbackLink = false

	// 9. Format standard CNI Result v1.0.0
	contIdx := 1
	res := &CNIResult{
		CNIVersion: CNIVersion,
		Interfaces: []ResultInterface{
			{Name: hostIfName, Mac: nk.Parent.Attrs().HardwareAddr.String()},
			{Name: env.IfName, Mac: podMac, Sandbox: env.NetNS},
		},
		IPs:    []ResultIPConfig{},
		Routes: []ResultRoute{},
	}

	if allocation.IPv4 != "" {
		gwStr := ""
		if allocation.Gateway4 != nil {
			gwStr = allocation.Gateway4.String()
		}
		res.IPs = append(res.IPs, ResultIPConfig{
			Interface: &contIdx,
			Address:   allocation.IPv4,
			Gateway:   gwStr,
		})
		res.Routes = append(res.Routes, ResultRoute{
			Dst: "0.0.0.0/0",
			GW:  gwStr,
		})
	}

	if allocation.IPv6 != "" {
		gwStr := ""
		if allocation.Gateway6 != nil {
			gwStr = allocation.Gateway6.String()
		}
		res.IPs = append(res.IPs, ResultIPConfig{
			Interface: &contIdx,
			Address:   allocation.IPv6,
			Gateway:   gwStr,
		})
		res.Routes = append(res.Routes, ResultRoute{
			Dst: "::/0",
			GW:  gwStr,
		})
	}

	return WriteResult(res)
}

