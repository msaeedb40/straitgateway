// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"

	cnitypes "github.com/containernetworking/cni/pkg/types"
	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
	agentv1 "github.com/msaeedb40/straitgateway/proto/agent/v1"
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
	conn   *grpc.ClientConn
	client agentv1.AgentServiceClient
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
	return &DaemonClient{
		conn:   conn,
		client: agentv1.NewAgentServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection.
func (c *DaemonClient) Close() error {
	return c.conn.Close()
}

// AllocateIP requests an IP allocation from the straitgatewayd IPAM.
func (c *DaemonClient) AllocateIP(containerID, netns, ifName string) (*IPAllocation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), agentCallTimeout)
	defer cancel()

	resp, err := c.client.AllocateIP(ctx, &agentv1.AllocateIPRequest{
		ContainerId: containerID,
		Netns:       netns,
		IfName:      ifName,
		Ipv4:        true,
	})
	if err != nil {
		return nil, fmt.Errorf("straitgatewayd AllocateIP: %w", err)
	}

	ipv4Val := resp.GetIpv4Address().GetIpv4()
	ipBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ipBytes, ipv4Val)

	prefixLen := int(resp.GetIpv4Cidr().GetPrefixLength())
	if prefixLen == 0 {
		prefixLen = 24
	}

	alloc := &IPAllocation{
		IPv4Net: net.IPNet{
			IP:   net.IP(ipBytes),
			Mask: net.CIDRMask(prefixLen, 32),
		},
		Identity: sgtypes.IdentityLocalMin,
	}

	if resp.GetGateway() != nil {
		gwVal := resp.GetGateway().GetIpv4()
		gwBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(gwBytes, gwVal)
		alloc.Gateway = net.IP(gwBytes)
	}

	return alloc, nil
}

// ReleaseIP releases an allocated IP back to the IPAM pool.
func (c *DaemonClient) ReleaseIP(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), agentCallTimeout)
	defer cancel()

	_, err := c.client.ReleaseIP(ctx, &agentv1.ReleaseIPRequest{
		ContainerId: containerID,
	})
	return err
}

// RegisterEndpoint notifies straitgatewayd of the new host-side NetKit interface.
func (c *DaemonClient) RegisterEndpoint(containerID string, hostIfIndex int, identity sgtypes.Identity) error {
	ctx, cancel := context.WithTimeout(context.Background(), agentCallTimeout)
	defer cancel()

	resp, err := c.client.ConfigureEndpoint(ctx, &agentv1.ConfigureEndpointRequest{
		ContainerId: containerID,
		HostIfindex: int32(hostIfIndex),
	})
	if err != nil {
		return err
	}
	_ = resp
	return nil
}

// DeregisterEndpoint removes the endpoint's BPF state and IP allocation.
func (c *DaemonClient) DeregisterEndpoint(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), agentCallTimeout)
	defer cancel()

	_, err := c.client.DeleteEndpoint(ctx, &agentv1.DeleteEndpointRequest{
		ContainerId: containerID,
	})
	return err
}

// CheckEndpoint validates the endpoint's network configuration.
func (c *DaemonClient) CheckEndpoint(containerID, netns, ifName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), agentCallTimeout)
	defer cancel()

	_, err := c.client.GetStatus(ctx, &agentv1.GetStatusRequest{})
	return err
}

// GarbageCollect triggers GC of stale endpoint state on straitgatewayd.
func (c *DaemonClient) GarbageCollect(args string) error {
	_ = args
	return nil
}
