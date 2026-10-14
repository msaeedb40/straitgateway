package datapath

import (
	"fmt"
	"sync"
	"time"
)

// PacingAlgorithm defines the active queue pacing algorithm
type PacingAlgorithm string

const (
	PacingEDT    PacingAlgorithm = "edt"    // Earliest Departure Time (FQ/EDT)
	PacingBBR    PacingAlgorithm = "bbr"    // BBR-based pacing
	PacingStatic PacingAlgorithm = "static" // Static configured rate
)

// FlowPacingState tracks live pacing parameters for an active L4 connection
type FlowPacingState struct {
	FlowID           string          `json:"flowId"`
	Algorithm        PacingAlgorithm `json:"algorithm"`
	CurrentRateBps   uint64          `json:"currentRateBps"`
	RTTMicroseconds  uint32          `json:"rttUs"`
	BDPBytes         uint64          `json:"bdpBytes"`
	ZeroWindowProbes uint32          `json:"zeroWindowProbes"`
	LastPacedAt      time.Time       `json:"lastPacedAt"`
}

// PacingConfig contains configuration for adaptive eBPF connection pacing
type PacingConfig struct {
	DefaultLinkRateBps uint64          `json:"defaultLinkRateBps"`
	TargetRTTUs        uint32          `json:"targetRttUs"`
	MinPacingRateBps   uint64          `json:"minPacingRateBps"`
	Algorithm          PacingAlgorithm `json:"algorithm"`
}

// PacingManager orchestrates kernel socket pacing and eBPF queue pacing
type PacingManager struct {
	mu     sync.RWMutex
	config PacingConfig
	flows  map[string]*FlowPacingState
}

// NewPacingManager creates a PacingManager
func NewPacingManager(cfg PacingConfig) *PacingManager {
	if cfg.DefaultLinkRateBps == 0 {
		cfg.DefaultLinkRateBps = 10 * 1000 * 1000 * 1000 // 10 Gbps default
	}
	if cfg.TargetRTTUs == 0 {
		cfg.TargetRTTUs = 200 // 200 microseconds data center RTT
	}
	if cfg.MinPacingRateBps == 0 {
		cfg.MinPacingRateBps = 100 * 1000 * 1000 // 100 Mbps min
	}
	if cfg.Algorithm == "" {
		cfg.Algorithm = PacingEDT
	}

	return &PacingManager{
		config: cfg,
		flows:  make(map[string]*FlowPacingState),
	}
}

// ComputePacingRate calculates the optimal bytes/sec rate based on BDP
func (pm *PacingManager) ComputePacingRate(rttUs uint32, observedBandwidthBps uint64) uint64 {
	if rttUs == 0 {
		rttUs = pm.config.TargetRTTUs
	}
	_ = rttUs // reserved for future BDP calculation
	effectiveBw := observedBandwidthBps
	if effectiveBw == 0 {
		effectiveBw = pm.config.DefaultLinkRateBps
	}

	// Calculate Bandwidth-Delay Product (BDP) in bits: Bw * RTT
	// Add 15% pacing headroom for smooth bursting without queue drop
	pacingRate := uint64(float64(effectiveBw) * 1.15)
	if pacingRate < pm.config.MinPacingRateBps {
		pacingRate = pm.config.MinPacingRateBps
	}
	return pacingRate
}

// RegisterFlow registers or updates an active TCP connection pacing state
func (pm *PacingManager) RegisterFlow(flowID string, rttUs uint32, observedBw uint64) *FlowPacingState {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	rate := pm.ComputePacingRate(rttUs, observedBw)
	bdpBytes := (rate / 8) * uint64(rttUs) / 1000000

	state := &FlowPacingState{
		FlowID:          flowID,
		Algorithm:       pm.config.Algorithm,
		CurrentRateBps:  rate,
		RTTMicroseconds: rttUs,
		BDPBytes:        bdpBytes,
		LastPacedAt:     time.Now(),
	}

	pm.flows[flowID] = state
	return state
}

// GetFlow returns current pacing parameters for a flow
func (pm *PacingManager) GetFlow(flowID string) (*FlowPacingState, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	st, exists := pm.flows[flowID]
	return st, exists
}

// RecordZeroWindow records zero-window probe suppression
func (pm *PacingManager) RecordZeroWindow(flowID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	st, exists := pm.flows[flowID]
	if !exists {
		return fmt.Errorf("flow %s not found", flowID)
	}

	st.ZeroWindowProbes++
	// Back off pacing rate by 50%
	st.CurrentRateBps = st.CurrentRateBps / 2
	if st.CurrentRateBps < pm.config.MinPacingRateBps {
		st.CurrentRateBps = pm.config.MinPacingRateBps
	}
	return nil
}

// ActiveFlowsCount returns number of monitored paced connections
func (pm *PacingManager) ActiveFlowsCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.flows)
}
