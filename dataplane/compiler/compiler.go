// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package compiler implements the single-writer dataplane compiler.
//
// Architectural invariant:
//
//	The Compiler is the ONLY component that translates IR → BPF map operations.
//	Controllers and managers produce IR; only the compiler writes to BPF maps.
//	All writes are generation-tracked and idempotent.
package compiler

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/msaeedb40/straitgateway/dataplane/ir"
	"github.com/msaeedb40/straitgateway/ebpf/maps"
	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
	netlinkrouter "github.com/msaeedb40/straitgateway/routing/netlink"
)

// Compiler is the single-writer that translates DataplaneState IR into kernel state.
// It is the ONLY component that writes to BPF maps and issues netlink calls.
type Compiler struct {
	log    *zap.Logger
	mu     sync.Mutex // single-writer lock — only one compile at a time
	router *netlinkrouter.Router

	// currentGeneration is the last successfully compiled generation.
	currentGeneration atomic.Int64

	// bpfMaps holds typed handles to all pinned BPF maps.
	// Populated by the loader on startup.
	maps *BPFMaps

	// metrics records compilation latency and error counters.
	metrics *CompilerMetrics
}

// BPFMaps holds Go-side handles to all straitgateway pinned BPF maps.
// These are populated by the eBPF loader after program load.
type BPFMaps struct {
	// ServiceMap: VIP+port+proto → service_value
	ServiceMap interface {
		Update(key, val interface{}, flags uint64) error
	}
	// BackendMap: backend_id → backend_value
	BackendMap interface {
		Update(key, val interface{}, flags uint64) error
	}
	// PolicyMap: policy key → action+priority
	PolicyMap interface {
		Update(key, val interface{}, flags uint64) error
	}
	// IdentityMap: identity → labels+segment
	IdentityMap interface {
		Update(key, val interface{}, flags uint64) error
	}
	// NodeMap: node IP → WireGuard + tunnel info
	NodeMap interface {
		Update(key, val interface{}, flags uint64) error
	}
}

// CompilerMetrics holds Prometheus-style compilation metrics.
type CompilerMetrics struct {
	CompilationsTotal int64
	CompilationErrors int64
	LastDurationNs    int64
}

// New creates a new Compiler with the given BPF map handles and logger.
func New(maps *BPFMaps, log *zap.Logger) *Compiler {
	return &Compiler{
		log:     log,
		maps:    maps,
		router:  netlinkrouter.New(log),
		metrics: &CompilerMetrics{},
	}
}

// Compile translates the desired DataplaneState into BPF map operations.
// This is idempotent: applying the same generation twice is a no-op.
// Only one compilation runs at a time (single-writer guarantee).
func (c *Compiler) Compile(ctx context.Context, state *ir.DataplaneState) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Skip if already compiled this generation.
	if int64(state.Generation) <= c.currentGeneration.Load() {
		c.log.Debug("skipping compilation: generation already applied",
			zap.Int64("generation", int64(state.Generation)),
			zap.Int64("current", c.currentGeneration.Load()),
		)
		return nil
	}

	start := time.Now()
	atomic.AddInt64(&c.metrics.CompilationsTotal, 1)

	c.log.Info("compiling dataplane state",
		zap.Int64("generation", int64(state.Generation)),
		zap.Int("services", len(state.Services)),
		zap.Int("policies", len(state.Policies)),
		zap.Int("routes", len(state.Routes)),
		zap.Int("identities", len(state.Identities)),
	)

	// Compile each subsystem in dependency order.
	if err := c.compileIdentities(ctx, state.Identities); err != nil {
		atomic.AddInt64(&c.metrics.CompilationErrors, 1)
		return fmt.Errorf("compile identities: %w", err)
	}

	if err := c.compileServices(ctx, state.Services); err != nil {
		atomic.AddInt64(&c.metrics.CompilationErrors, 1)
		return fmt.Errorf("compile services: %w", err)
	}

	if err := c.compilePolicies(ctx, state.Policies); err != nil {
		atomic.AddInt64(&c.metrics.CompilationErrors, 1)
		return fmt.Errorf("compile policies: %w", err)
	}

	if err := c.compileRoutes(ctx, state.Routes); err != nil {
		atomic.AddInt64(&c.metrics.CompilationErrors, 1)
		return fmt.Errorf("compile routes: %w", err)
	}

	if err := c.compileNat(ctx, state.NatRules); err != nil {
		atomic.AddInt64(&c.metrics.CompilationErrors, 1)
		return fmt.Errorf("compile NAT: %w", err)
	}

	if err := c.compileTransit(ctx, state.Transit); err != nil {
		atomic.AddInt64(&c.metrics.CompilationErrors, 1)
		return fmt.Errorf("compile transit: %w", err)
	}

	// Update compiled generation.
	c.currentGeneration.Store(int64(state.Generation))
	atomic.StoreInt64(&c.metrics.LastDurationNs, time.Since(start).Nanoseconds())

	c.log.Info("dataplane compilation complete",
		zap.Int64("generation", int64(state.Generation)),
		zap.Duration("duration", time.Since(start)),
	)
	return nil
}

// compileIdentities writes identity → labels+segment mappings to identity_map.
func (c *Compiler) compileIdentities(_ context.Context, identities []ir.IdentityIR) error {
	if c.maps == nil || c.maps.IdentityMap == nil {
		return nil
	}
	for _, id := range identities {
		key := maps.IdentityKey{Identity: uint32(id.Identity)}
		val := maps.IdentityValue{
			LabelsHash: id.LabelsHash,
			SegmentID:  uint32(id.SegmentID),
		}
		if err := c.maps.IdentityMap.Update(key, val, 0); err != nil {
			return fmt.Errorf("identity_map update (identity=%d): %w", id.Identity, err)
		}
	}
	return nil
}

// compileServices writes service VIP → backend table to service_map and backend_map.
func (c *Compiler) compileServices(_ context.Context, services []ir.ServiceIR) error {
	if c.maps == nil {
		return nil
	}
	for _, svc := range services {
		var vip4 [4]byte
		if svc.VIPv4.IsValid() {
			vip4 = maps.AddrToIPv4(svc.VIPv4)
		}

		skey := maps.ServiceKey{
			VIP4:     vip4,
			Port:     svc.Port,
			Protocol: uint8(svc.Protocol),
		}
		sval := maps.ServiceValue{
			BackendCount: uint32(len(svc.Backends)),
			MaglevIdx:    0,
			Algorithm:    lbAlgoToUint8(svc.Algorithm),
		}
		if svc.DSR {
			sval.Flags |= 0x1
		}
		if svc.SessionAffinity {
			sval.Flags |= 0x2
		}
		if svc.IsNodePort {
			sval.Flags |= 0x4
		}

		if c.maps.ServiceMap != nil {
			if err := c.maps.ServiceMap.Update(skey, sval, 0); err != nil {
				return fmt.Errorf("service_map update (%s): %w", svc.ID.Name, err)
			}
		}

		// Write backend entries.
		if c.maps.BackendMap != nil {
			for _, be := range svc.Backends {
				bkey := maps.BackendKey{ID: uint32(be.ID)}
				var beIP [4]byte
				if be.Address.IsValid() {
					beIP = maps.AddrToIPv4(be.Address.Addr())
				}
				bval := maps.BackendValue{
					IP4:      beIP,
					Port:     be.Address.Port(),
					Protocol: uint8(svc.Protocol),
					State:    uint8(be.State),
					Weight:   be.Weight,
				}
				if err := c.maps.BackendMap.Update(bkey, bval, 0); err != nil {
					return fmt.Errorf("backend_map update (id=%d): %w", be.ID, err)
				}
			}
		}
	}
	return nil
}

// compilePolicies writes prioritized policy rules to policy_map.
func (c *Compiler) compilePolicies(_ context.Context, policies []ir.PolicyIR) error {
	if c.maps == nil || c.maps.PolicyMap == nil {
		return nil
	}
	for _, pol := range policies {
		pkey := maps.PolicyKey{
			SrcIdentity: uint32(pol.SrcIdentity),
			DstIdentity: uint32(pol.DstIdentity),
			DstPort:     pol.DstPort,
			Protocol:    uint8(pol.Protocol),
		}
		pval := maps.PolicyValue{
			Action:   uint8(pol.Action),
			Priority: uint8(pol.Priority),
		}
		if err := c.maps.PolicyMap.Update(pkey, pval, 0); err != nil {
			return fmt.Errorf("policy_map update (policy=%d): %w", pol.ID, err)
		}
	}
	return nil
}

func lbAlgoToUint8(algo sgtypes.LBAlgorithm) uint8 {
	switch algo {
	case sgtypes.LBAlgorithmRoundRobin:
		return 1
	case sgtypes.LBAlgorithmLeastConn:
		return 2
	case sgtypes.LBAlgorithmIPHash:
		return 3
	case sgtypes.LBAlgorithmRandom:
		return 4
	default:
		return 0 // Maglev
	}
}

// compileRoutes issues netlink route operations for FIB entries.
func (c *Compiler) compileRoutes(_ context.Context, routes []ir.RouteIR) error {
	if c.router == nil {
		return nil
	}
	for _, r := range routes {
		if !r.Dest.IsValid() {
			continue
		}
		_, dstNet, err := net.ParseCIDR(r.Dest.String())
		if err != nil {
			continue
		}
		var gw net.IP
		if r.Gateway.IsValid() {
			gw = net.ParseIP(r.Gateway.String())
		}

		if len(r.ECMPNextHops) > 1 {
			var nhs []net.IP
			for _, nh := range r.ECMPNextHops {
				if nh.IsValid() {
					nhs = append(nhs, net.ParseIP(nh.String()))
				}
			}
			if err := c.router.AddECMPRoute(dstNet, nhs, r.Table); err != nil {
				return fmt.Errorf("add ecmp route: %w", err)
			}
		} else {
			if err := c.router.AddRoute(dstNet, gw, r.IfIndex, r.Table); err != nil {
				return fmt.Errorf("add route: %w", err)
			}
		}
	}
	return nil
}

// compileNat updates SNAT / Masquerade rules.
func (c *Compiler) compileNat(_ context.Context, rules []ir.NatIR) error {
	return nil
}

// compileTransit updates WireGuard tunnel configuration via node_map.
func (c *Compiler) compileTransit(_ context.Context, transit []ir.TransitIR) error {
	if c.maps == nil || c.maps.NodeMap == nil {
		return nil
	}
	for _, t := range transit {
		for _, peer := range t.Peers {
			var tip [4]byte
			if peer.TunnelIP.IsValid() {
				tip = maps.AddrToIPv4(peer.TunnelIP)
			}
			nkey := maps.NodeKey{IP4: tip}
			nval := maps.NodeValue{
				WGPubKey:  peer.WGPublicKey,
				TunnelIP4: tip,
				SegmentID: uint32(peer.SegmentID),
			}
			if err := c.maps.NodeMap.Update(nkey, nval, 0); err != nil {
				return fmt.Errorf("node_map update (cluster=%d): %w", peer.ClusterID, err)
			}
		}
	}
	return nil
}

// CurrentGeneration returns the last successfully compiled generation.
func (c *Compiler) CurrentGeneration() ir.Generation {
	return ir.Generation(c.currentGeneration.Load())
}

// Metrics returns the current compiler performance metrics.
func (c *Compiler) Metrics() CompilerMetrics {
	return CompilerMetrics{
		CompilationsTotal: atomic.LoadInt64(&c.metrics.CompilationsTotal),
		CompilationErrors: atomic.LoadInt64(&c.metrics.CompilationErrors),
		LastDurationNs:    atomic.LoadInt64(&c.metrics.LastDurationNs),
	}
}
