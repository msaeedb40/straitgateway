// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package main implements the straitgateway CNI plugin binary.
//
// Implements CNI Spec 1.1+: ADD, DEL, CHECK, GC, VERSION operations.
//
// Architectural invariants enforced here:
//   - CNI ADD synchronous fast-path NEVER blocks on Service/Policy/NAT/Gateway/Transit/BGP
//   - API server reachable via node routing BEFORE the CNI is called
//   - Service and Policy reconciliation are async (decoupled from ADD return)
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
)

const pluginName = "straitgateway"

// NetConf is the CNI network configuration passed by the runtime.
type NetConf struct {
	types.NetConf
	// SocketPath is the straitgatewayd Unix socket for IPAM gRPC calls.
	SocketPath string `json:"socketPath"`
	// MTU overrides auto-detected MTU. 0 = auto.
	MTU int `json:"mtu,omitempty"`
	// EnableIPv4 enables IPv4 pod networking. Default: true.
	EnableIPv4 bool `json:"enableIPv4"`
	// EnableIPv6 enables IPv6 pod networking. Default: false.
	EnableIPv6 bool `json:"enableIPv6"`
}

func main() {
	skel.PluginMainFuncs(
		skel.CNIFuncs{
			Add:   cmdAdd,
			Del:   cmdDel,
			Check: cmdCheck,
			GC:    cmdGC,
		},
		version.PluginSupports("0.3.0", "0.3.1", "0.4.0", "1.0.0", "1.1.0"),
		pluginName,
	)
}

// cmdAdd implements CNI ADD — allocates IP, creates NetKit, configures routes.
//
// Synchronous fast-path ONLY:
//  1. Parse config
//  2. Allocate IP from straitgatewayd IPAM (gRPC, fast local call)
//  3. Create NetKit device pair
//  4. Configure interface and address
//  5. Configure mandatory routes (node routing, NOT via Service LB)
//  6. Allocate BPF identity
//  7. Return CNI result
//
// Service map, Policy, NAT, Observability are updated asynchronously by
// straitgatewayd AFTER returning the CNI result.
func cmdAdd(args *skel.CmdArgs) error {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return fmt.Errorf("straitgateway CNI ADD: parse config: %w", err)
	}

	// 1. Contact straitgatewayd for IPAM + identity allocation.
	client, err := newDaemonClient(conf.SocketPath)
	if err != nil {
		return fmt.Errorf("straitgateway CNI ADD: dial daemon: %w", err)
	}
	defer func() { _ = client.Close() }()

	alloc, err := client.AllocateIP(args.ContainerID, args.Netns, args.IfName)
	if err != nil {
		return fmt.Errorf("straitgateway CNI ADD: allocate IP: %w", err)
	}

	// 2 + 3. Create NetKit device pair and move container side into netns.
	hostIface, err := setupNetKit(args.Netns, args.IfName, alloc, conf.MTU)
	if err != nil {
		// Best-effort cleanup on failure
		_ = client.ReleaseIP(args.ContainerID)
		return fmt.Errorf("straitgateway CNI ADD: setup NetKit: %w", err)
	}

	// 4 + 5. Configure routes inside the pod network namespace.
	if err := configureRoutes(args.Netns, args.IfName, alloc); err != nil {
		_ = client.ReleaseIP(args.ContainerID)
		return fmt.Errorf("straitgateway CNI ADD: configure routes: %w", err)
	}

	// 6. Notify daemon of host-side ifindex for BPF identity + map setup.
	// This triggers async dataplane reconciliation (service map, policy, NAT).
	if err := client.RegisterEndpoint(args.ContainerID, hostIface.Index, alloc.Identity); err != nil {
		log.Printf("straitgateway CNI ADD: register endpoint (non-fatal): %v", err)
		// Non-fatal: async reconciliation will retry
	}

	// 7. Build and return CNI result.
	result := &current.Result{
		CNIVersion: conf.CNIVersion,
		Interfaces: []*current.Interface{
			{
				Name:    args.IfName,
				Sandbox: args.Netns,
			},
			{
				Name: hostIface.Name,
			},
		},
		IPs: []*current.IPConfig{
			{
				Interface: current.Int(0),
				Address:   alloc.IPv4Net,
				Gateway:   alloc.Gateway,
			},
		},
		Routes: alloc.Routes,
		DNS:    conf.DNS,
	}

	return types.PrintResult(result, conf.CNIVersion)
}

// cmdDel implements CNI DEL — releases IP, removes routes, destroys NetKit.
//
// Sequence (reverse of ADD):
//  1. Remove routes
//  2. Deregister BPF identity
//  3. Release IP back to IPAM
//  4. Destroy NetKit device pair
func cmdDel(args *skel.CmdArgs) error {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return fmt.Errorf("straitgateway CNI DEL: parse config: %w", err)
	}

	client, err := newDaemonClient(conf.SocketPath)
	if err != nil {
		// Daemon may be restarting — best effort cleanup
		log.Printf("straitgateway CNI DEL: dial daemon (non-fatal): %v", err)
		return nil
	}
	defer func() { _ = client.Close() }()

	// Deregister endpoint (removes BPF identity, policy state, service map entry)
	if err := client.DeregisterEndpoint(args.ContainerID); err != nil {
		log.Printf("straitgateway CNI DEL: deregister (non-fatal): %v", err)
	}

	// Release IP
	if err := client.ReleaseIP(args.ContainerID); err != nil {
		log.Printf("straitgateway CNI DEL: release IP (non-fatal): %v", err)
	}

	// Destroy NetKit — removing the host-side interface automatically
	// removes the peer container-side interface.
	if err := teardownNetKit(args.ContainerID); err != nil {
		log.Printf("straitgateway CNI DEL: teardown NetKit (non-fatal): %v", err)
	}

	return nil
}

// cmdCheck implements CNI CHECK — validates the network configuration is still correct.
func cmdCheck(args *skel.CmdArgs) error {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return fmt.Errorf("straitgateway CNI CHECK: parse config: %w", err)
	}

	client, err := newDaemonClient(conf.SocketPath)
	if err != nil {
		return fmt.Errorf("straitgateway CNI CHECK: dial daemon: %w", err)
	}
	defer func() { _ = client.Close() }()

	return client.CheckEndpoint(args.ContainerID, args.Netns, args.IfName)
}

// cmdGC implements CNI GC — garbage collects stale endpoint state.
func cmdGC(args *skel.CmdArgs) error {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return fmt.Errorf("straitgateway CNI GC: parse config: %w", err)
	}

	client, err := newDaemonClient(conf.SocketPath)
	if err != nil {
		return fmt.Errorf("straitgateway CNI GC: dial daemon: %w", err)
	}
	defer func() { _ = client.Close() }()

	return client.GarbageCollect(args.Args)
}

func parseConfig(stdin []byte) (*NetConf, error) {
	conf := &NetConf{
		SocketPath: "/run/straitgateway/daemon.sock",
		EnableIPv4: true,
	}
	if err := json.Unmarshal(stdin, conf); err != nil {
		return nil, fmt.Errorf("parsing CNI config: %w", err)
	}
	return conf, nil
}

func init() {
	// Ensure we never accidentally inherit file descriptors.
	// CNI plugins must be careful about open FDs.
	if err := os.Stderr.Close(); err == nil {
		os.Stderr, _ = os.Open(os.DevNull)
	}
}
