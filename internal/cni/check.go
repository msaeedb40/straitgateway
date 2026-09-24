package cni

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/straitgateway/straitgateway/internal/cni/ipam"
	"github.com/straitgateway/straitgateway/internal/linux/namespace"
	"github.com/straitgateway/straitgateway/internal/tgwd/api"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

// CmdCheck verifies that the pod's network configuration matches expected state.
func CmdCheck(stdin io.Reader) error {
	var netConf NetConf
	if err := json.NewDecoder(stdin).Decode(&netConf); err != nil {
		return fmt.Errorf("cni: failed to parse netconf: %w", err)
	}

	env, err := ParseCNIEnv()
	if err != nil {
		return fmt.Errorf("cni: failed to parse environment: %w", err)
	}

	if env.NetNS == "" || env.ContainerID == "" {
		return fmt.Errorf("cni: CNI_NETNS and CNI_CONTAINERID are required for CHECK")
	}

	// 1. Verify host interface exists
	hostIfName := HostInterfaceName(env.ContainerID)
	if _, err := netlink.LinkByName(hostIfName); err != nil {
		return fmt.Errorf("cni check: host interface %s not found: %w", hostIfName, err)
	}

	// 2. Verify IPAM record
	store := ipam.NewStore("")
	alloc, exists, err := store.Get(env.ContainerID)
	if err != nil || !exists {
		return fmt.Errorf("cni check: no IPAM lease found for container %s", env.ContainerID)
	}

	// 3. Verify pod netns and container interface
	ns, err := namespace.GetFromPath(env.NetNS)
	if err != nil {
		return fmt.Errorf("cni check: unable to open pod netns %s: %w", env.NetNS, err)
	}
	defer ns.Close()

	if err := ns.Do(func() error {
		link, err := netlink.LinkByName(env.IfName)
		if err != nil {
			return fmt.Errorf("cni check: container interface %s missing: %w", env.IfName, err)
		}
		if link.Attrs().OperState == netlink.OperDown {
			return fmt.Errorf("cni check: container interface %s is DOWN", env.IfName)
		}

		addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
		if err != nil {
			return fmt.Errorf("cni check: list addrs on %s: %w", env.IfName, err)
		}

		matchedV4 := alloc.IPv4 == ""
		matchedV6 := alloc.IPv6 == ""

		for _, addr := range addrs {
			addrStr := addr.IPNet.String()
			if alloc.IPv4 != "" && addrStr == alloc.IPv4 {
				matchedV4 = true
			}
			if alloc.IPv6 != "" && addrStr == alloc.IPv6 {
				matchedV6 = true
			}
		}

		if !matchedV4 {
			return fmt.Errorf("cni check: configured IPv4 %s not found on interface %s", alloc.IPv4, env.IfName)
		}
		if !matchedV6 {
			return fmt.Errorf("cni check: configured IPv6 %s not found on interface %s", alloc.IPv6, env.IfName)
		}

		return nil
	}); err != nil {
		return err
	}

	// 4. Verify eBPF datapath state (if StraitD is reachable)
	socketPath := ResolveSocketPath(netConf.StraitDSocket)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	client, errClient := api.NewClient(ctx, zap.NewNop(), socketPath)
	if errClient == nil {
		defer client.Close()
		state, err := client.GetNodeState(ctx)
		if err == nil && !state.DatapathReady {
			return fmt.Errorf("cni check: datapath not ready on node")
		}
	}

	return nil
}
