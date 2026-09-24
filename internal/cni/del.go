package cni

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/straitgateway/straitgateway/internal/cni/ipam"
	"github.com/straitgateway/straitgateway/internal/linux/link"
	"github.com/straitgateway/straitgateway/internal/tgwd/api"
	"go.uber.org/zap"
)

// CmdDel executes the CNI DEL operation. DEL is strictly idempotent.
func CmdDel(stdin io.Reader) error {
	var netConf NetConf
	_ = json.NewDecoder(stdin).Decode(&netConf)

	env, err := ParseCNIEnv()
	if err != nil || env.ContainerID == "" {
		// DEL must never fail when containerID is missing
		return nil
	}

	hostIfName := HostInterfaceName(env.ContainerID)

	// 1. Notify StraitD (if running)
	socketPath := ResolveSocketPath(netConf.StraitDSocket)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	client, errClient := api.NewClient(ctx, zap.NewNop(), socketPath)
	if errClient == nil {
		defer client.Close()
		_, _ = client.DeletePodNetwork(ctx, &api.DeletePodNetworkRequest{
			Header: &api.ResourceHeader{
				ResourceUid: env.ContainerID,
			},
			ContainerId: env.ContainerID,
			NetnsPath:   env.NetNS,
		})
	}

	// 2. Remove host-side interface (and its peer if still linked)
	_ = link.DeleteNetkit(hostIfName)

	// 3. Release IPAM allocation
	store := ipam.NewStore("")
	alloc, exists, _ := store.Get(env.ContainerID)
	if exists {
		cidrs := make([]string, 0)
		for _, r := range netConf.IPAM.Ranges {
			if r.Subnet != "" {
				cidrs = append(cidrs, r.Subnet)
			}
		}
		if len(cidrs) == 0 && env.ParsedArgs.PodCIDR != "" {
			for _, part := range strings.Split(env.ParsedArgs.PodCIDR, ",") {
				if strings.TrimSpace(part) != "" {
					cidrs = append(cidrs, strings.TrimSpace(part))
				}
			}
		}
		if len(cidrs) > 0 {
			allocator, errAlloc := ipam.NewAllocator(cidrs)
			if errAlloc == nil {
				if alloc.IPv4 != "" {
					_ = allocator.Release(alloc.IPv4)
				}
				if alloc.IPv6 != "" {
					_ = allocator.Release(alloc.IPv6)
				}
			}
		}
		_ = store.Delete(env.ContainerID)
	}

	return nil
}
