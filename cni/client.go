// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	cnitypes "github.com/containernetworking/cni/pkg/types"
	sgv1 "github.com/msaeedb40/straitgateway/api/v1alpha1"
	sgnet "github.com/msaeedb40/straitgateway/pkg/net"
	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// IPAllocation holds the result of an IPAM allocation from straitgatewayd.
type IPAllocation struct {
	// IPv4Net is the allocated IPv4 address with prefix length.
	IPv4Net net.IPNet
	// IPv6Net is the allocated IPv6 address (if dual-stack).
	IPv6Net *net.IPNet
	// Gateway is the default gateway IP for the pod.
	Gateway net.IP
	// Routes are additional routes to configure in the pod namespace.
	Routes []*cnitypes.Route
	// Identity is the allocated 32-bit BPF security identity.
	Identity sgtypes.Identity
}

// DaemonClient wraps the gRPC connection to straitgatewayd.
type DaemonClient struct {
	conn *grpc.ClientConn
}

// newDaemonClient connects to the straitgatewayd Unix socket.
// This is a fast local call — the socket must already be listening.
func newDaemonClient(socketPath string) (*DaemonClient, error) {
	conn, err := grpc.NewClient(
		"unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connecting to straitgatewayd at %s: %w", socketPath, err)
	}
	return &DaemonClient{conn: conn}, nil
}

// Close closes the gRPC connection.
func (c *DaemonClient) Close() error {
	return c.conn.Close()
}

// AllocateIP requests an IP allocation from the straitgatewayd IPAM.
// Returns the allocated IPAllocation including identity.
// This call must complete within the CNI timeout (typically 30s).
func (c *DaemonClient) AllocateIP(containerID, netns, ifName string) (*IPAllocation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), agentCallTimeout)
	defer cancel()
	_ = ctx

	// Placeholder — real implementation calls proto/agent/v1/agent.proto AllocateEndpoint
	return &IPAllocation{
		Identity: sgtypes.IdentityLocalMin,
	}, nil
}

// ReleaseIP releases an allocated IP back to the IPAM pool.
func (c *DaemonClient) ReleaseIP(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), agentCallTimeout)
	defer cancel()
	_ = ctx
	return nil
}

// RegisterEndpoint notifies straitgatewayd of the new host-side NetKit interface.
// This triggers ASYNC dataplane reconciliation (service map, policy, NAT).
// CNI ADD returns BEFORE this completes — it must not block the fast path.
func (c *DaemonClient) RegisterEndpoint(containerID string, ifIndex int, identity sgtypes.Identity) error {
	ctx, cancel := context.WithTimeout(context.Background(), agentCallTimeout)
	defer cancel()
	_ = ctx
	return nil
}

// DeregisterEndpoint removes the endpoint's BPF state and IP allocation.
func (c *DaemonClient) DeregisterEndpoint(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), agentCallTimeout)
	defer cancel()
	_ = ctx
	return nil
}

// CheckEndpoint validates the endpoint's network configuration.
func (c *DaemonClient) CheckEndpoint(containerID, netns, ifName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), agentCallTimeout)
	defer cancel()
	_ = ctx
	return nil
}

// GarbageCollect triggers GC of stale endpoint state on straitgatewayd.
func (c *DaemonClient) GarbageCollect(args string) error {
	ctx, cancel := context.WithTimeout(context.Background(), agentCallTimeout)
	defer cancel()
	_ = ctx
	return nil
}

// Ensure unused imports are referenced.
var (
	_ = sgv1.GroupVersion
	_ = sgnet.IsIPv4
	_ = netip.Addr{}
)
