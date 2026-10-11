// Package roce — dcqcn.go implements the DCQCN (Data Center Quantized
// Congestion Notification) rate controller for StraitGateway RoCEv2.
//
// DCQCN is the standard congestion control algorithm for RoCEv2, defined in:
//   - "DCQCN: Congestion Control for Large-Scale RDMA Deployments" (SIGCOMM 2015)
//   - RFC 8257 (ECN for TCP) — adapted concepts for RDMA
//   - IBTA Annex 16A (RoCEv2 Congestion Management)
//
// # Algorithm Overview
//
// DCQCN uses a two-sided protocol:
//
//  1. Receiver side (StraitGateway XDP): when ECN-CE is detected on a packet,
//     the receiver generates a Congestion Notification Packet (CNP) toward
//     the sender. Our BPF program emits an event to sg_roce_events; userspace
//     sends the CNP via rdma-core.
//
//  2. Sender side (this Go module): when a CNP is received, apply the DCQCN
//     rate-reduction algorithm:
//     - Reduce sending rate by factor α (default 1/2)
//     - Update Additive Increase counter
//     - Set the BC (Byte Counter) timer
//
// # Rate Controller
//
// The RateController tracks per-QPN sending rate and applies:
//   - Rate reduction on CNP: Rc ← Rc × (1 - α/2), where α ∈ [0,1]
//   - Additive increase on no-CNP intervals: Rc ← Rc + RAI (Rate Additive Inc)
//   - Hyperactive increase (HAI): if no CNP for K timer expirations, Rc ← Rc + RHAI
//   - Rate is clamped to [MinRateMbps, MaxRateMbps]
//
// The actual enforcement is via RDMA QP modify (ibv_modify_qp with rate limit),
// delegated to the rdma-core integration layer.
package roce

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// DCQCNConfig holds the DCQCN algorithm parameters.
// These match the commonly deployed defaults from Mellanox/NVIDIA RDMA stacks.
type DCQCNConfig struct {
	// Alpha is the initial ECN marking fraction (0.0–1.0). Default: 1.0.
	Alpha float64
	// AlphaDecay controls how fast Alpha decreases on non-CNP intervals.
	// Formula: α_new = (1-g) × α_old + g × F  where F = 0 if no CE. Default: 0.0625.
	AlphaDecay float64
	// RateDecreaseInterval is the minimum interval between rate decreases (default: 4ms).
	RateDecreaseInterval time.Duration
	// RateAI is the additive increase step in Mbps per timer expiration. Default: 5 Mbps.
	RateAI float64
	// RateHAI is the hyperactive increase step in Mbps. Default: 50 Mbps.
	RateHAI float64
	// MinRateMbps is the floor rate in Mbps. Default: 100 Mbps.
	MinRateMbps float64
	// MaxRateMbps is the ceiling rate in Mbps. Default: 100000 Mbps (100 Gbps).
	MaxRateMbps float64
	// ByteCounterThreshold triggers rate increase after N bytes without CNP. Default: 150000 bytes.
	ByteCounterThreshold uint64
	// TimerPeriod is the DCQCN timer granularity. Default: 55ms (IBTA default).
	TimerPeriod time.Duration
	// HAITimerCount is the number of non-CNP timer expirations before HAI. Default: 5.
	HAITimerCount int
}

// DefaultDCQCNConfig returns the default DCQCN parameters.
func DefaultDCQCNConfig() DCQCNConfig {
	return DCQCNConfig{
		Alpha:                1.0,
		AlphaDecay:           0.0625,
		RateDecreaseInterval: 4 * time.Millisecond,
		RateAI:               5.0,
		RateHAI:              50.0,
		MinRateMbps:          100.0,
		MaxRateMbps:          100_000.0,
		ByteCounterThreshold: 150_000,
		TimerPeriod:          55 * time.Millisecond,
		HAITimerCount:        5,
	}
}

// QPRateState holds the per-QPN DCQCN rate control state.
type QPRateState struct {
	// CurrentRateMbps is the current allowed sending rate.
	CurrentRateMbps float64
	// TargetRateMbps is the rate we are converging toward.
	TargetRateMbps float64
	// Alpha is the current ECN fraction estimate.
	Alpha float64
	// ByteCounter counts bytes sent since last rate increase.
	ByteCounter uint64
	// TimerCount counts consecutive non-CNP timer expirations (for HAI).
	TimerCount int
	// LastCNPTime is when the last CNP was received for this QPN.
	LastCNPTime time.Time
	// LastDecreaseTime is when the rate was last decreased.
	LastDecreaseTime time.Time
	// TotalCNPs is the total number of CNPs received.
	TotalCNPs uint64
}

// RateController implements the DCQCN per-QPN rate controller.
type RateController struct {
	mu  sync.Mutex
	cfg DCQCNConfig
	qps map[uint32]*QPRateState

	// OnRateChange is called when a QPN's rate changes.
	// The caller should apply the new rate via RDMA QP modify.
	OnRateChange func(qpn uint32, newRateMbps float64)
}

// NewRateController creates a DCQCN rate controller with the given config.
func NewRateController(cfg DCQCNConfig) *RateController {
	return &RateController{
		cfg: cfg,
		qps: make(map[uint32]*QPRateState),
	}
}

// RegisterQP initialises rate state for a new QP at the maximum rate.
func (rc *RateController) RegisterQP(qpn uint32) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.qps[qpn] = &QPRateState{
		CurrentRateMbps: rc.cfg.MaxRateMbps,
		TargetRateMbps:  rc.cfg.MaxRateMbps,
		Alpha:           rc.cfg.Alpha,
	}
}

// DeregisterQP removes rate state for a completed/destroyed QP.
func (rc *RateController) DeregisterQP(qpn uint32) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	delete(rc.qps, qpn)
}

// OnCNPReceived applies the DCQCN rate reduction when a CNP is received.
// This implements the receiver-side CNP processing described in SIGCOMM 2015 §3.
func (rc *RateController) OnCNPReceived(qpn uint32) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	st, ok := rc.qps[qpn]
	if !ok {
		return
	}

	now := time.Now()

	// Rate decrease throttle — don't decrease faster than RateDecreaseInterval
	if now.Sub(st.LastDecreaseTime) < rc.cfg.RateDecreaseInterval {
		return
	}

	st.TotalCNPs++
	st.LastCNPTime = now
	st.LastDecreaseTime = now
	st.TimerCount = 0 // reset HAI counter

	// Update alpha: α_new = (1-g)×α + g×1 (CE fraction → 1 on CNP)
	st.Alpha = (1-rc.cfg.AlphaDecay)*st.Alpha + rc.cfg.AlphaDecay*1.0
	if st.Alpha > 1.0 {
		st.Alpha = 1.0
	}

	// Rate reduction: Rc_new = Rc × (1 - α/2)
	newRate := st.CurrentRateMbps * (1.0 - st.Alpha/2.0)
	newRate = math.Max(newRate, rc.cfg.MinRateMbps)

	// Save current rate as target (for slow start)
	st.TargetRateMbps = st.CurrentRateMbps
	st.CurrentRateMbps = newRate
	st.ByteCounter = 0

	rc.notifyRateChange(qpn, newRate)
}

// OnTimerExpiry is called on every DCQCN timer tick for a QPN. If no CNP was
// received in this period, the rate is increased (AI or HAI mode).
func (rc *RateController) OnTimerExpiry(qpn uint32, bytesSent uint64) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	st, ok := rc.qps[qpn]
	if !ok {
		return
	}

	st.ByteCounter += bytesSent

	// Check if we should increase the rate
	if time.Since(st.LastCNPTime) < rc.cfg.TimerPeriod {
		// CNP was recent — decay Alpha toward 0 (no congestion)
		st.Alpha = (1 - rc.cfg.AlphaDecay) * st.Alpha
		return
	}

	// No CNP in last timer period — apply AI or HAI
	st.TimerCount++
	var increase float64
	if st.TimerCount >= rc.cfg.HAITimerCount {
		increase = rc.cfg.RateHAI
	} else {
		increase = rc.cfg.RateAI
	}

	// Additive increase toward target rate, then beyond
	newRate := st.CurrentRateMbps + increase
	if newRate > rc.cfg.MaxRateMbps {
		newRate = rc.cfg.MaxRateMbps
	}

	// Byte counter threshold: additional fast recovery
	if st.ByteCounter >= rc.cfg.ByteCounterThreshold {
		newRate = math.Min(newRate+increase, rc.cfg.MaxRateMbps)
		st.ByteCounter = 0
	}

	// Decay Alpha toward 0 (less congestion signal)
	st.Alpha = (1 - rc.cfg.AlphaDecay) * st.Alpha

	st.CurrentRateMbps = newRate
	rc.notifyRateChange(qpn, newRate)
}

// QPStats returns a snapshot of the rate state for a given QPN.
func (rc *RateController) QPStats(qpn uint32) (QPRateState, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	st, ok := rc.qps[qpn]
	if !ok {
		return QPRateState{}, fmt.Errorf("QPN %d not registered", qpn)
	}
	return *st, nil
}

// AllQPStats returns a snapshot of all registered QPN rate states.
func (rc *RateController) AllQPStats() map[uint32]QPRateState {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	out := make(map[uint32]QPRateState, len(rc.qps))
	for qpn, st := range rc.qps {
		out[qpn] = *st
	}
	return out
}

func (rc *RateController) notifyRateChange(qpn uint32, rateMbps float64) {
	if rc.OnRateChange != nil {
		rc.OnRateChange(qpn, rateMbps)
	}
}
