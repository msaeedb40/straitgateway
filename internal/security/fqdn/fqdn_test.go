package fqdn

import (
	"net"
	"testing"
	"time"
)

func TestFQDNRulePatternMatching(t *testing.T) {
	exactRule := &Rule{
		Pattern:   "api.github.com",
		MatchType: MatchExact,
	}

	if !exactRule.Matches("api.github.com") {
		t.Errorf("exact rule should match api.github.com")
	}
	if !exactRule.Matches("API.GITHUB.COM.") {
		t.Errorf("exact rule should match case-insensitive with trailing dot")
	}
	if exactRule.Matches("gist.github.com") {
		t.Errorf("exact rule should not match gist.github.com")
	}

	wildcardRule := &Rule{
		Pattern:   "*.amazonaws.com",
		MatchType: MatchWildcard,
	}

	if !wildcardRule.Matches("s3.amazonaws.com") {
		t.Errorf("wildcard rule should match s3.amazonaws.com")
	}
	if !wildcardRule.Matches("ec2.us-east-1.amazonaws.com") {
		t.Errorf("wildcard rule should match ec2.us-east-1.amazonaws.com")
	}
	if wildcardRule.Matches("evil-amazonaws.com") {
		t.Errorf("wildcard rule should not match evil-amazonaws.com")
	}
}

func TestFQDNDNSSnooperAndLeaseExpiry(t *testing.T) {
	engine := NewEngine()

	// 1. Register wildcard rule
	_, err := engine.RegisterRule("*.openai.com", "ALLOW")
	if err != nil {
		t.Fatalf("failed to register rule: %v", err)
	}

	testIP := net.ParseIP("104.18.6.192")

	// 2. Process DNS response
	entries, matched := engine.ProcessDNSResponse("api.openai.com.", []net.IP{testIP}, 1) // 1 second TTL
	if !matched || len(entries) != 1 {
		t.Fatalf("expected matched DNS response with 1 entry")
	}

	// 3. Verify IP is allowed immediately
	if !engine.IsIPAllowed(testIP) {
		t.Errorf("expected test IP to be authorized")
	}

	// 4. Wait for TTL expiry
	time.Sleep(1100 * time.Millisecond)

	// 5. Verify IP is now expired
	if engine.IsIPAllowed(testIP) {
		t.Errorf("expected test IP to expire after 1s TTL")
	}

	// 6. Prune expired entries
	pruned := engine.PruneExpiredEntries()
	if pruned != 1 {
		t.Errorf("expected 1 entry pruned, got %d", pruned)
	}
	if engine.ActiveEntriesCount() != 0 {
		t.Errorf("expected 0 active entries after pruning")
	}
}
