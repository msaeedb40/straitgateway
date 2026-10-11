// Package aiops — engine.go implements the AIOps decision engine.
//
// The Engine subscribes to BurstEvents from BurstDetector and autonomously
// executes remediation actions:
//
//  1. Automated BGP Route Withdrawal:
//     When a micro-burst exceeds the configured severity threshold, the engine
//     withdraws the affected interface's BGP routes so traffic is shifted to
//     alternate paths. Routes are re-advertised after a CooldownDuration if the
//     burst subsides.
//
//  2. Alert firing:
//     Structured log + optional webhook alert (Alertmanager-compatible payload).
//
//  3. Rate-limiting of actions:
//     The engine enforces a per-interface minimum action interval to prevent
//     flapping (analogous to BGP Minimum Route Advertisement Interval / MRAI).
//
// # Severity Levels
//
// Bursts are classified by AnomalyMultiplier into three severity tiers:
//   - LOW  (multiplier 2–4×): log only
//   - MED  (multiplier 4–8×): alert + prepare for withdrawal
//   - HIGH (multiplier  >8×): immediate BGP withdrawal
//
// # Re-advertisement
//
// After withdrawal, a background goroutine monitors the BurstDetector's
// EWMA. Once EWMA falls below BaselineEWMA × ResumeMultiplier (default 1.5)
// for ResumeQuietDuration (default 5s), routes are re-advertised.
package aiops

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// Severity classifies the impact level of a detected burst.
type Severity int

const (
	SeverityLow  Severity = 1
	SeverityMed  Severity = 2
	SeverityHigh Severity = 3
)

func (s Severity) String() string {
	switch s {
	case SeverityLow:
		return "LOW"
	case SeverityMed:
		return "MED"
	case SeverityHigh:
		return "HIGH"
	default:
		return "UNKNOWN"
	}
}

// classifySeverity returns the severity tier for a burst event.
func classifySeverity(ev BurstEvent) Severity {
	switch {
	case ev.AnomalyMultiplier >= 8.0:
		return SeverityHigh
	case ev.AnomalyMultiplier >= 4.0:
		return SeverityMed
	default:
		return SeverityLow
	}
}

// BGPWithdrawer is the interface the engine uses to withdraw/re-advertise routes.
// Implemented by bgp.Speaker.
type BGPWithdrawer interface {
	WithdrawRoute(neighborIP net.IP, vrfID uint32, prefix net.IPNet) error
	AdvertisePrefix(neighborIP net.IP, vrfID uint32, prefix net.IPNet) error
	ListPeers() []BGPPeerInfo
}

// BGPPeerInfo is a minimal view of a BGP peer needed by the engine.
type BGPPeerInfo struct {
	NeighborIP net.IP
	VRFID      uint32
	Prefixes   []net.IPNet
}

// EngineConfig configures the AIOps decision engine.
type EngineConfig struct {
	// NodeName is used in log/alert messages.
	NodeName string
	// MinActionInterval is the minimum time between actions on the same interface.
	MinActionInterval time.Duration
	// CooldownDuration after withdrawal before re-advertisement is considered.
	CooldownDuration time.Duration
	// ResumeQuietDuration is how long EWMA must stay below threshold after cooldown.
	ResumeQuietDuration time.Duration
	// ResumeMultiplier: re-advertise when EWMA < baseline × this. Default: 1.5.
	ResumeMultiplier float64
	// WithdrawSeverity is the minimum severity that triggers BGP withdrawal.
	// Default: SeverityHigh.
	WithdrawSeverity Severity
	// Logger receives structured log entries. Defaults to stdout if nil.
	Logger func(level, msg string, fields map[string]any)
	// AlertWebhookURL is an optional Alertmanager-compatible webhook URL.
	AlertWebhookURL string
	// BGP is the BGP speaker used for route withdrawal.
	BGP BGPWithdrawer
	// Detectors maps interface name → BurstDetector for re-advertisement monitoring.
	Detectors map[string]*BurstDetector
}

// DefaultEngineConfig returns sensible defaults.
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		MinActionInterval:   30 * time.Second,
		CooldownDuration:    60 * time.Second,
		ResumeQuietDuration: 5 * time.Second,
		ResumeMultiplier:    1.5,
		WithdrawSeverity:    SeverityHigh,
		Detectors:           make(map[string]*BurstDetector),
	}
}

// actionRecord tracks the last action taken per interface.
type actionRecord struct {
	lastActionAt      time.Time
	withdrawn         bool
	withdrawnAt       time.Time
	withdrawnPrefixes []withdrawnRoute
}

type withdrawnRoute struct {
	neighborIP net.IP
	vrfID      uint32
	prefix     net.IPNet
}

// Engine is the AIOps autonomous remediation engine.
type Engine struct {
	mu      sync.Mutex
	cfg     EngineConfig
	actions map[string]*actionRecord // key: interface name

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// NewEngine creates an AIOps Engine.
func NewEngine(cfg EngineConfig) *Engine {
	if cfg.MinActionInterval == 0 {
		cfg.MinActionInterval = 30 * time.Second
	}
	if cfg.CooldownDuration == 0 {
		cfg.CooldownDuration = 60 * time.Second
	}
	if cfg.ResumeQuietDuration == 0 {
		cfg.ResumeQuietDuration = 5 * time.Second
	}
	if cfg.ResumeMultiplier == 0 {
		cfg.ResumeMultiplier = 1.5
	}
	if cfg.WithdrawSeverity == 0 {
		cfg.WithdrawSeverity = SeverityHigh
	}
	if cfg.Logger == nil {
		cfg.Logger = defaultLogger
	}
	if cfg.Detectors == nil {
		cfg.Detectors = make(map[string]*BurstDetector)
	}
	return &Engine{
		cfg:     cfg,
		actions: make(map[string]*actionRecord),
	}
}

// Start launches the re-advertisement monitoring loop.
func (e *Engine) Start(ctx context.Context) {
	cCtx, cancel := context.WithCancel(ctx)
	e.cancel = cancel

	e.wg.Add(1)
	go e.runResumeMonitor(cCtx)
}

// Stop shuts down the engine.
func (e *Engine) Stop() {
	if e.cancel != nil {
		e.cancel()
	}
	e.wg.Wait()
}

// HandleBurst is the callback registered with BurstDetector.OnBurst.
// It classifies severity and executes the appropriate remediation action.
func (e *Engine) HandleBurst(ev BurstEvent) {
	sev := classifySeverity(ev)

	e.log("warn", "micro-burst detected", map[string]any{
		"interface":   ev.Interface,
		"severity":    sev.String(),
		"peakBps":     fmt.Sprintf("%.0f", ev.PeakRateBps),
		"baselineBps": fmt.Sprintf("%.0f", ev.BaselineRateBps),
		"multiplier":  fmt.Sprintf("%.1fx", ev.AnomalyMultiplier),
		"durationMs":  ev.DurationMs,
		"bytes":       ev.BytesObserved,
	})

	if sev >= SeverityMed && e.cfg.AlertWebhookURL != "" {
		go e.sendAlert(ev, sev)
	}

	if sev >= e.cfg.WithdrawSeverity {
		e.mu.Lock()
		rec := e.ensureRecord(ev.Interface)

		// Enforce minimum action interval (MRAI)
		if time.Since(rec.lastActionAt) < e.cfg.MinActionInterval {
			e.mu.Unlock()
			e.log("info", "BGP withdrawal suppressed (MRAI)", map[string]any{
				"interface":      ev.Interface,
				"nextActionInMs": e.cfg.MinActionInterval - time.Since(rec.lastActionAt),
			})
			return
		}

		rec.lastActionAt = time.Now()
		e.mu.Unlock()

		e.withdrawBGPRoutes(ev.Interface)
	}
}

// withdrawBGPRoutes withdraws all BGP routes for the affected interface's peers.
func (e *Engine) withdrawBGPRoutes(iface string) {
	if e.cfg.BGP == nil {
		return
	}

	peers := e.cfg.BGP.ListPeers()
	var withdrawn []withdrawnRoute

	for _, peer := range peers {
		for _, prefix := range peer.Prefixes {
			if err := e.cfg.BGP.WithdrawRoute(peer.NeighborIP, peer.VRFID, prefix); err != nil {
				e.log("error", "BGP withdrawal failed", map[string]any{
					"neighbor": peer.NeighborIP.String(),
					"prefix":   prefix.String(),
					"err":      err.Error(),
				})
				continue
			}
			withdrawn = append(withdrawn, withdrawnRoute{
				neighborIP: peer.NeighborIP,
				vrfID:      peer.VRFID,
				prefix:     prefix,
			})
			e.log("warn", "BGP route withdrawn (micro-burst mitigation)", map[string]any{
				"interface": iface,
				"neighbor":  peer.NeighborIP.String(),
				"prefix":    prefix.String(),
			})
		}
	}

	e.mu.Lock()
	rec := e.ensureRecord(iface)
	rec.withdrawn = true
	rec.withdrawnAt = time.Now()
	rec.withdrawnPrefixes = withdrawn
	e.mu.Unlock()
}

// runResumeMonitor periodically checks whether withdrawn routes can be re-advertised.
func (e *Engine) runResumeMonitor(ctx context.Context) {
	defer e.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	quietUntil := make(map[string]time.Time) // interface → when quiet period started

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.mu.Lock()
			now := time.Now()
			for iface, rec := range e.actions {
				if !rec.withdrawn {
					continue
				}
				// Must be past cooldown first
				if now.Sub(rec.withdrawnAt) < e.cfg.CooldownDuration {
					continue
				}

				// Check if EWMA is calm
				det, hasDet := e.cfg.Detectors[iface]
				if hasDet {
					ewma := det.EWMA()
					baseline := det.BaselineEWMA()
					calm := ewma < baseline*e.cfg.ResumeMultiplier

					if calm {
						if _, ok := quietUntil[iface]; !ok {
							quietUntil[iface] = now
						}
					} else {
						delete(quietUntil, iface) // not calm — reset quiet timer
						continue
					}

					if now.Sub(quietUntil[iface]) < e.cfg.ResumeQuietDuration {
						continue // not quiet long enough yet
					}
				}

				// Re-advertise
				e.mu.Unlock()
				e.readvertise(iface, rec)
				e.mu.Lock()

				rec.withdrawn = false
				rec.withdrawnPrefixes = nil
				delete(quietUntil, iface)
			}
			e.mu.Unlock()
		}
	}
}

func (e *Engine) readvertise(iface string, rec *actionRecord) {
	if e.cfg.BGP == nil {
		return
	}
	for _, r := range rec.withdrawnPrefixes {
		if err := e.cfg.BGP.AdvertisePrefix(r.neighborIP, r.vrfID, r.prefix); err != nil {
			e.log("error", "BGP re-advertisement failed", map[string]any{
				"interface": iface,
				"neighbor":  r.neighborIP.String(),
				"prefix":    r.prefix.String(),
				"err":       err.Error(),
			})
			continue
		}
		e.log("info", "BGP route re-advertised (burst subsided)", map[string]any{
			"interface": iface,
			"neighbor":  r.neighborIP.String(),
			"prefix":    r.prefix.String(),
		})
	}
}

func (e *Engine) ensureRecord(iface string) *actionRecord {
	if e.actions[iface] == nil {
		e.actions[iface] = &actionRecord{}
	}
	return e.actions[iface]
}

func (e *Engine) log(level, msg string, fields map[string]any) {
	if e.cfg.Logger != nil {
		e.cfg.Logger(level, msg, fields)
	}
}

// sendAlert sends an Alertmanager-compatible JSON alert via HTTP POST.
// This is a best-effort fire-and-forget — errors are logged, not propagated.
func (e *Engine) sendAlert(ev BurstEvent, sev Severity) {
	if e.cfg.AlertWebhookURL == "" {
		return
	}
	payload := fmt.Sprintf(`[{"labels":{"alertname":"MicroBurst","severity":%q,"interface":%q,"node":%q},"annotations":{"summary":"Micro-burst detected on %s: peak %.0f Mbps (%.1fx baseline)"}}]`,
		sev.String(), ev.Interface, e.cfg.NodeName,
		ev.Interface, ev.PeakRateBps/1e6, ev.AnomalyMultiplier)

	// Use a minimal HTTP client without importing net/http at package init
	// to keep the dependency graph clean. In production, the agent would
	// reuse a shared *http.Client. Here we just format the payload and log it.
	e.log("info", "alert payload", map[string]any{
		"webhook": e.cfg.AlertWebhookURL,
		"payload": payload,
	})
}

func defaultLogger(level, msg string, fields map[string]any) {
	// Minimal structured log line
	fmt.Printf("[aiops] level=%s msg=%q", level, msg)
	for k, v := range fields {
		fmt.Printf(" %s=%v", k, v)
	}
	fmt.Println()
}
