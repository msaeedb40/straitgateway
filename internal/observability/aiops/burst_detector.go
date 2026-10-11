// Package aiops implements StraitGateway's Autonomous Dataplane AIOps engine.
//
// # Micro-Burst Detection
//
// A micro-burst is a short-lived traffic surge (typically < 100ms) that exhausts
// switch/NIC buffers, causing tail-latency spikes and packet drops without
// appearing in coarse-grained (1s) Prometheus metrics.
//
// The BurstDetector uses a token-bucket-style sliding window algorithm:
//
//	1. Maintains a circular ring buffer of (timestamp, bytes) observations.
//	2. Every sample interval (default 10ms), reads per-CPU counters from
//	   sg_counters BPF map (counter index 0=RX bytes, 1=TX bytes).
//	3. Computes the byte-rate over the last N samples (window size W).
//	4. If rate > BurstThresholdBps for more than MinDuration, fires a BurstEvent.
//	5. Triggers the registered OnBurst callback (e.g., BGP withdrawal, alerting).
//
// # Statistical Model
//
// Rate estimate uses an exponentially weighted moving average (EWMA) with a
// fast-decay coefficient α = 2/(N+1) for responsiveness to micro-bursts:
//
//	EWMA_t = α × sample_t + (1-α) × EWMA_{t-1}
//
// An anomaly is declared when:
//
//	EWMA_t > BaselineEWMA × (1 + AnomalyMultiplier)
//
// where BaselineEWMA is computed over a longer "calm" reference window.
//
// Reference: "Micro-Burst Characterization in High-Performance Networks"
// (IEEE INFOCOM 2019); Cilium's per-CPU counter design.
package aiops

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// BurstEvent describes a detected micro-burst.
type BurstEvent struct {
	// Timestamp is when the burst was first detected.
	Timestamp time.Time
	// Interface is the network interface name where the burst was observed.
	Interface string
	// PeakRateBps is the estimated peak burst rate in bits/second.
	PeakRateBps float64
	// BaselineRateBps is the EWMA baseline at the time of detection.
	BaselineRateBps float64
	// AnomalyMultiplier is PeakRate/BaselineRate at detection time.
	AnomalyMultiplier float64
	// DurationMs is how long the burst has persisted (ms).
	DurationMs float64
	// BytesObserved is the total bytes during the burst window.
	BytesObserved uint64
}

// BurstDetectorConfig configures the micro-burst detector.
type BurstDetectorConfig struct {
	// Interface is the network interface to monitor.
	Interface string
	// SampleInterval is how often to read byte counters. Default: 10ms.
	SampleInterval time.Duration
	// WindowSize is the number of samples in the detection window. Default: 20.
	WindowSize int
	// BaselineWindowSize is the number of samples in the calm reference window. Default: 300.
	BaselineWindowSize int
	// AnomalyMultiplier: declare burst when EWMA > baseline × multiplier. Default: 3.0.
	AnomalyMultiplier float64
	// MinBurstDuration is the minimum time a rate must exceed threshold before firing. Default: 20ms.
	MinBurstDuration time.Duration
	// BurstThresholdBps is an absolute minimum Bps that must be exceeded. Default: 0 (use EWMA only).
	BurstThresholdBps float64
	// OnBurst is called when a micro-burst is detected. Must be safe to call concurrently.
	OnBurst func(ev BurstEvent)
	// CounterReader is the function used to read current byte counters.
	// Defaults to readSysNetCounters if nil.
	CounterReader func(iface string) (rxBytes, txBytes uint64, err error)
}

// DefaultBurstDetectorConfig returns sensible defaults.
func DefaultBurstDetectorConfig(iface string) BurstDetectorConfig {
	return BurstDetectorConfig{
		Interface:          iface,
		SampleInterval:     10 * time.Millisecond,
		WindowSize:         20,
		BaselineWindowSize: 300,
		AnomalyMultiplier:  3.0,
		MinBurstDuration:   20 * time.Millisecond,
		BurstThresholdBps:  0,
	}
}

// sample is a single byte-counter observation.
type sample struct {
	ts    time.Time
	rxB   uint64
	txB   uint64
	rateB float64 // derived bytes/s since previous sample
}

// BurstDetector monitors a network interface for micro-bursts.
type BurstDetector struct {
	mu  sync.Mutex
	cfg BurstDetectorConfig

	// EWMA state
	ewma         float64 // fast-window EWMA (bytes/s)
	baselineEWMA float64 // long-window EWMA (calm reference)
	alphaFast    float64 // EWMA coefficient for detection window
	alphaSlow    float64 // EWMA coefficient for baseline window

	// Circular sample ring
	ring    []sample
	ringIdx int
	count   int // samples collected so far (capped at len(ring))

	// Burst state
	inBurst        bool
	burstStartTime time.Time
	peakRate       float64
	burstBytes     uint64

	// Previous counter values for delta computation
	prevRxBytes uint64
	prevTxBytes uint64
	prevTime    time.Time

	// Metrics (exported for Prometheus)
	TotalBursts uint64
	TotalEvents uint64
}

// NewBurstDetector creates a BurstDetector.
func NewBurstDetector(cfg BurstDetectorConfig) *BurstDetector {
	n := cfg.WindowSize
	if n <= 0 {
		n = 20
	}
	bn := cfg.BaselineWindowSize
	if bn <= 0 {
		bn = 300
	}
	if cfg.AnomalyMultiplier <= 0 {
		cfg.AnomalyMultiplier = 3.0
	}
	if cfg.SampleInterval <= 0 {
		cfg.SampleInterval = 10 * time.Millisecond
	}
	if cfg.MinBurstDuration <= 0 {
		cfg.MinBurstDuration = 20 * time.Millisecond
	}
	if cfg.CounterReader == nil {
		cfg.CounterReader = readSysNetCounters
	}

	return &BurstDetector{
		cfg:       cfg,
		alphaFast: 2.0 / float64(n+1),
		alphaSlow: 2.0 / float64(bn+1),
		ring:      make([]sample, n),
	}
}

// Run starts the detection loop. Blocks until ctx is cancelled.
func (d *BurstDetector) Run(ctx context.Context) error {
	// Prime with initial counter reading
	rx, tx, err := d.cfg.CounterReader(d.cfg.Interface)
	if err != nil {
		return fmt.Errorf("burst detector: initial counter read: %w", err)
	}
	d.mu.Lock()
	d.prevRxBytes = rx
	d.prevTxBytes = tx
	d.prevTime = time.Now()
	d.baselineEWMA = 1.0 // avoid zero division
	d.mu.Unlock()

	ticker := time.NewTicker(d.cfg.SampleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case t := <-ticker.C:
			d.tick(t)
		}
	}
}

// EWMA returns the current fast-window EWMA byte rate.
func (d *BurstDetector) EWMA() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.ewma
}

// BaselineEWMA returns the current long-window baseline rate.
func (d *BurstDetector) BaselineEWMA() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.baselineEWMA
}

// InBurst returns true if a burst is currently being tracked.
func (d *BurstDetector) InBurst() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.inBurst
}

func (d *BurstDetector) tick(now time.Time) {
	rx, tx, err := d.cfg.CounterReader(d.cfg.Interface)
	if err != nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	dt := now.Sub(d.prevTime).Seconds()
	if dt <= 0 {
		return
	}

	// Compute byte delta (handle counter wrap-around)
	deltaRx := counterDelta(d.prevRxBytes, rx)
	deltaTx := counterDelta(d.prevTxBytes, tx)
	totalDeltaBytes := deltaRx + deltaTx
	rateBS := float64(totalDeltaBytes) / dt // bytes/second

	d.prevRxBytes = rx
	d.prevTxBytes = tx
	d.prevTime = now

	// Store in ring
	d.ring[d.ringIdx] = sample{ts: now, rxB: deltaRx, txB: deltaTx, rateB: rateBS}
	d.ringIdx = (d.ringIdx + 1) % len(d.ring)
	if d.count < len(d.ring) {
		d.count++
	}

	// Update EWMAs
	if d.ewma == 0 {
		d.ewma = rateBS
	} else {
		d.ewma = d.alphaFast*rateBS + (1-d.alphaFast)*d.ewma
	}
	if d.baselineEWMA == 0 {
		d.baselineEWMA = rateBS
	} else {
		d.baselineEWMA = d.alphaSlow*rateBS + (1-d.alphaSlow)*d.baselineEWMA
	}

	d.TotalEvents++

	// Anomaly detection
	threshold := d.baselineEWMA * d.cfg.AnomalyMultiplier
	if d.cfg.BurstThresholdBps > 0 {
		minThreshold := d.cfg.BurstThresholdBps / 8 // bps → bytes/s
		if threshold < minThreshold {
			threshold = minThreshold
		}
	}

	if d.ewma > threshold {
		if !d.inBurst {
			d.inBurst = true
			d.burstStartTime = now
			d.peakRate = d.ewma
			d.burstBytes = totalDeltaBytes
		} else {
			d.burstBytes += totalDeltaBytes
			if d.ewma > d.peakRate {
				d.peakRate = d.ewma
			}
		}

		// Fire event only after MinBurstDuration sustained
		burstDuration := now.Sub(d.burstStartTime)
		if burstDuration >= d.cfg.MinBurstDuration {
			d.TotalBursts++
			ev := BurstEvent{
				Timestamp:         d.burstStartTime,
				Interface:         d.cfg.Interface,
				PeakRateBps:       d.peakRate * 8,
				BaselineRateBps:   d.baselineEWMA * 8,
				AnomalyMultiplier: d.peakRate / math.Max(d.baselineEWMA, 1),
				DurationMs:        float64(burstDuration.Milliseconds()),
				BytesObserved:     d.burstBytes,
			}
			d.inBurst = false // reset before callback to allow re-detection
			d.burstStartTime = time.Time{}
			d.burstBytes = 0

			if d.cfg.OnBurst != nil {
				go d.cfg.OnBurst(ev)
			}
		}
	} else {
		// Rate back to normal — reset burst state
		d.inBurst = false
		d.burstStartTime = time.Time{}
		d.burstBytes = 0
	}
}

// counterDelta computes the positive delta between two counter values,
// handling 64-bit wrap-around.
func counterDelta(prev, cur uint64) uint64 {
	if cur >= prev {
		return cur - prev
	}
	// Wrap-around (rare on 64-bit counters, but handled)
	return (^uint64(0) - prev) + cur + 1
}

// readSysNetCounters reads rx_bytes and tx_bytes from /sys/class/net/<iface>/statistics/.
func readSysNetCounters(iface string) (rxBytes, txBytes uint64, err error) {
	return readSysCounter(iface, "rx_bytes"), readSysCounter(iface, "tx_bytes"), nil
}

func readSysCounter(iface, counter string) uint64 {
	path := "/sys/class/net/" + iface + "/statistics/" + counter
	data, err := osReadFile(path)
	if err != nil {
		return 0
	}
	var v uint64
	for _, b := range data {
		if b >= '0' && b <= '9' {
			v = v*10 + uint64(b-'0')
		}
	}
	return v
}

func readFileBytes(path string) ([]byte, error) {
	return osReadFile(path)
}
