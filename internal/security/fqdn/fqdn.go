package fqdn

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// MatchType defines domain pattern matching style
type MatchType string

const (
	MatchExact    MatchType = "exact"
	MatchWildcard MatchType = "wildcard"
)

// Rule defines an egress domain security rule
type Rule struct {
	ID        string    `json:"id"`
	Pattern   string    `json:"pattern"`
	MatchType MatchType `json:"matchType"`
	Action    string    `json:"action"` // "ALLOW", "DENY"
}

// ResolvedEntry holds an active DNS resolution bound to an eBPF policy
type ResolvedEntry struct {
	IP        net.IP    `json:"ip"`
	Domain    string    `json:"domain"`
	RuleID    string    `json:"ruleId"`
	TTL       uint32    `json:"ttl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Engine manages FQDN egress rules and DNS response snooping
type Engine struct {
	mu          sync.RWMutex
	rules       map[string]*Rule
	resolvedIPs map[string]*ResolvedEntry // key: IP string
}

// NewEngine creates an FQDN egress security engine
func NewEngine() *Engine {
	return &Engine{
		rules:       make(map[string]*Rule),
		resolvedIPs: make(map[string]*ResolvedEntry),
	}
}

// RegisterRule adds an FQDN policy rule
func (e *Engine) RegisterRule(pattern string, action string) (*Rule, error) {
	if pattern == "" {
		return nil, fmt.Errorf("domain pattern cannot be empty")
	}

	matchType := MatchExact
	if strings.HasPrefix(pattern, "*.") {
		matchType = MatchWildcard
	}

	rule := &Rule{
		ID:        fmt.Sprintf("fqdn-%s", pattern),
		Pattern:   strings.ToLower(pattern),
		MatchType: matchType,
		Action:    strings.ToUpper(action),
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules[rule.ID] = rule
	return rule, nil
}

// Matches checks if a domain satisfies an FQDN rule pattern
func (r *Rule) Matches(domain string) bool {
	d := strings.ToLower(strings.TrimSuffix(domain, "."))
	if r.MatchType == MatchExact {
		return d == r.Pattern
	}
	if r.MatchType == MatchWildcard {
		suffix := strings.TrimPrefix(r.Pattern, "*.")
		return strings.HasSuffix(d, "."+suffix) || d == suffix
	}
	return false
}

// ProcessDNSResponse intercepts DNS A/AAAA responses and dynamically authorizes IPs
func (e *Engine) ProcessDNSResponse(domain string, ips []net.IP, ttlSeconds uint32) ([]*ResolvedEntry, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	domainClean := strings.ToLower(strings.TrimSuffix(domain, "."))
	var matchedRule *Rule

	for _, r := range e.rules {
		if r.Matches(domainClean) {
			matchedRule = r
			break
		}
	}

	if matchedRule == nil {
		return nil, false
	}

	if ttlSeconds == 0 {
		ttlSeconds = 60 // Minimum safety TTL
	}

	entries := make([]*ResolvedEntry, 0, len(ips))
	expiry := time.Now().Add(time.Duration(ttlSeconds) * time.Second)

	for _, ip := range ips {
		entry := &ResolvedEntry{
			IP:        ip,
			Domain:    domainClean,
			RuleID:    matchedRule.ID,
			TTL:       ttlSeconds,
			ExpiresAt: expiry,
		}
		e.resolvedIPs[ip.String()] = entry
		entries = append(entries, entry)
	}

	return entries, true
}

// IsIPAllowed queries if a destination IP has an active valid FQDN egress lease
func (e *Engine) IsIPAllowed(ip net.IP) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	entry, exists := e.resolvedIPs[ip.String()]
	if !exists {
		return false
	}

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		return false
	}
	return true
}

// PruneExpiredEntries garbage collects stale DNS entries from the eBPF policy cache
func (e *Engine) PruneExpiredEntries() int {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	pruned := 0

	for ipStr, entry := range e.resolvedIPs {
		if now.After(entry.ExpiresAt) {
			delete(e.resolvedIPs, ipStr)
			pruned++
		}
	}
	return pruned
}

// ActiveEntriesCount returns number of dynamically authorized destination IPs
func (e *Engine) ActiveEntriesCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.resolvedIPs)
}
