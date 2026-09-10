// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package compiler implements the single-writer dataplane compiler.
//
// Architectural invariant:
//   The Compiler is the ONLY component that translates IR → BPF map operations.
//   Controllers and managers produce IR; only the compiler writes to BPF maps.
//   All writes are generation-tracked and idempotent.
package compiler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/msaeedb40/straitgateway/dataplane/ir"
)

// Compiler is the single-writer that translates DataplaneState IR into kernel state.
// It is the ONLY component that writes to BPF maps and issues netlink calls.
type Compiler struct {
	log     *zap.Logger
	mu      sync.Mutex // single-writer lock — only one compile at a time

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
	ServiceMap interface{ Update(key, val interface{}, flags uint64) error }
	// BackendMap: backend_id → backend_value
	BackendMap interface{ Update(key, val interface{}, flags uint64) error }
	// PolicyMap: policy key → action+priority
	PolicyMap interface{ Update(key, val interface{}, flags uint64) error }
	// IdentityMap: identity → labels+segment
	IdentityMap interface{ Update(key, val interface{}, flags uint64) error }
	// NodeMap: node IP → WireGuard + tunnel info
	NodeMap interface{ Update(key, val interface{}, flags uint64) error }
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
	for _, id := range identities {
		key := map[string]interface{}{"identity": uint32(id.Identity)}
		val := map[string]interface{}{
			"labels_hash": id.LabelsHash,
			"segment_id":  uint32(id.SegmentID),
		}
		if err := c.maps.IdentityMap.Update(key, val, 0); err != nil {
			return fmt.Errorf("identity_map update (identity=%d): %w", id.Identity, err)
		}
	}
	return nil
}

// compileServices writes service VIP → backend table to service_map and backend_map.
func (c *Compiler) compileServices(_ context.Context, services []ir.ServiceIR) error {
	for _, svc := range services {
		// Write service entry.
		skey := map[string]interface{}{
			"vip4":     uint32(0), // TODO: convert svc.VIPv4 to uint32
			"port":     svc.Port,
			"protocol": uint8(svc.Protocol),
		}
		sval := map[string]interface{}{
			"backend_count": uint32(len(svc.Backends)),
			"algorithm":     uint8(0), // 0=maglev
		}
		if err := c.maps.ServiceMap.Update(skey, sval, 0); err != nil {
			return fmt.Errorf("service_map update (%s): %w", svc.ID.Name, err)
		}

		// Write backend entries.
		for _, be := range svc.Backends {
			bkey := map[string]interface{}{"id": uint32(be.ID)}
			bval := map[string]interface{}{
				"ip4":      uint32(0), // TODO: convert be.Address.Addr()
				"port":     be.Address.Port(),
				"protocol": uint8(svc.Protocol),
				"state":    uint8(be.State),
				"weight":   be.Weight,
			}
			if err := c.maps.BackendMap.Update(bkey, bval, 0); err != nil {
				return fmt.Errorf("backend_map update (id=%d): %w", be.ID, err)
			}
		}
	}
	return nil
}

// compilePolicies writes prioritized policy rules to policy_map.
func (c *Compiler) compilePolicies(_ context.Context, policies []ir.PolicyIR) error {
	for _, pol := range policies {
		pkey := map[string]interface{}{
			"src_identity": uint32(pol.SrcIdentity),
			"dst_identity": uint32(pol.DstIdentity),
			"dst_port":     pol.DstPort,
			"protocol":     uint8(pol.Protocol),
		}
		pval := map[string]interface{}{
			"action":   uint8(pol.Action),
			"priority": uint32(pol.Priority),
		}
		if err := c.maps.PolicyMap.Update(pkey, pval, 0); err != nil {
			return fmt.Errorf("policy_map update (policy=%d): %w", pol.ID, err)
		}
	}
	return nil
}

// compileRoutes issues netlink route operations for FIB entries.
// Routes are applied via the netlink API, not BPF maps.
func (c *Compiler) compileRoutes(_ context.Context, routes []ir.RouteIR) error {
	// TODO: implement netlink route apply via routing/netlink package.
	return nil
}

// compileNat updates SNAT configuration in the snat_config BPF array map.
func (c *Compiler) compileNat(_ context.Context, rules []ir.NatIR) error {
	// TODO: implement NAT rule compilation to BPF ct_map / snat_config.
	return nil
}

// compileTransit updates WireGuard tunnel configuration via node_map.
func (c *Compiler) compileTransit(_ context.Context, transit []ir.TransitIR) error {
	for _, t := range transit {
		for _, peer := range t.Peers {
			nkey := map[string]interface{}{
				"ip4": uint32(0), // TODO: convert peer.TunnelIP
			}
			nval := map[string]interface{}{
				"wg_pubkey":  peer.WGPublicKey,
				"tunnel_ip4": uint32(0),
				"segment_id": uint32(peer.SegmentID),
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
