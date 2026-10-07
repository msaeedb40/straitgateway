package mesh

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
)

// CircuitBreakerState represents health status of a backend
type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "Closed"
	StateOpen     CircuitBreakerState = "Open"
	StateHalfOpen CircuitBreakerState = "HalfOpen"
)

// CircuitBreakerConfig defines threshold and timeouts
type CircuitBreakerConfig struct {
	FailureThreshold int
	CooldownPeriod   time.Duration
}

// CircuitBreaker tracks failure rates and trips on consecutive errors
type CircuitBreaker struct {
	mu               sync.Mutex
	state            CircuitBreakerState
	consecutiveFails int
	lastFailedAt     time.Time
	config           CircuitBreakerConfig
}

// NewCircuitBreaker creates a circuit breaker
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.CooldownPeriod <= 0 {
		cfg.CooldownPeriod = 10 * time.Second
	}
	return &CircuitBreaker{
		state:  StateClosed,
		config: cfg,
	}
}

// Allow checks if a request is permitted
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	if cb.state == StateOpen {
		if now.Sub(cb.lastFailedAt) >= cb.config.CooldownPeriod {
			cb.state = StateHalfOpen
			return true
		}
		return false
	}
	return true
}

// RecordSuccess clears error counts and closes circuit
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.consecutiveFails = 0
	cb.state = StateClosed
}

// RecordFailure increments error counts and trips breaker if threshold exceeded
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.consecutiveFails++
	cb.lastFailedAt = time.Now()
	if cb.consecutiveFails >= cb.config.FailureThreshold {
		cb.state = StateOpen
	}
}

// State returns current state
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// FaultInjectionConfig specifies chaos policies
type FaultInjectionConfig struct {
	AbortPercentage float64
	AbortStatusCode int
	DelayPercentage float64
	DelayDuration   time.Duration
}

// CanaryHeaderRule routes to specific backend on matching header
type CanaryHeaderRule struct {
	HeaderName    string
	HeaderValue   string
	TargetBackend string
}

// ProxyRoute holds L7 routing and traffic split rules
type ProxyRoute struct {
	ServiceRef     string
	TrafficSplit   []v1alpha1.TrafficSplitTarget
	Retries        *v1alpha1.RetryPolicy
	Timeout        time.Duration
	MTLSMode       v1alpha1.MTLSMode
	CanaryRules    []CanaryHeaderRule
	CircuitBreaker *CircuitBreaker
	FaultInjection *FaultInjectionConfig
	Filters        *FilterPipeline
}

// NodeProxy is a shared node-level L7 proxy for sidecarless mesh traffic
type NodeProxy struct {
	mu         sync.RWMutex
	routes     map[string]*ProxyRoute // key: host/service
	httpClient *http.Client
}

// NewNodeProxy creates a NodeProxy
func NewNodeProxy() *NodeProxy {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Mock or custom CA in live clusters
		},
	}

	return &NodeProxy{
		routes: make(map[string]*ProxyRoute),
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

// RegisterRoute adds or updates an L7 route policy
func (p *NodeProxy) RegisterRoute(route *ProxyRoute) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.routes[route.ServiceRef] = route
}

// SelectDestination chooses a backend based on traffic split weights
func (p *NodeProxy) SelectDestination(serviceRef string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	route, exists := p.routes[serviceRef]
	if !exists || len(route.TrafficSplit) == 0 {
		return serviceRef, nil // Default to origin
	}

	totalWeight := int32(0)
	for _, target := range route.TrafficSplit {
		totalWeight += target.Weight
	}
	if totalWeight <= 0 {
		return route.TrafficSplit[0].BackendService, nil
	}

	choice := rand.Int31n(totalWeight)
	cumulative := int32(0)
	for _, target := range route.TrafficSplit {
		cumulative += target.Weight
		if choice < cumulative {
			return target.BackendService, nil
		}
	}
	return route.TrafficSplit[0].BackendService, nil
}

// ExecuteRequest forwards a request with retries, filters, circuit breaking, and fault injection
func (p *NodeProxy) ExecuteRequest(req *http.Request, serviceRef string) (*http.Response, error) {
	p.mu.RLock()
	route := p.routes[serviceRef]
	p.mu.RUnlock()

	// 1. Circuit Breaker Check
	if route != nil && route.CircuitBreaker != nil {
		if !route.CircuitBreaker.Allow() {
			return nil, fmt.Errorf("circuit breaker open: backend unavailable")
		}
	}

	// 2. L7 Filter Pipeline Execution
	var fCtx *FilterContext
	if route != nil && route.Filters != nil {
		fCtx = &FilterContext{
			Request:       req,
			Headers:       req.Header.Clone(),
			Attributes:    make(map[string]interface{}),
			TargetService: serviceRef,
			Timestamp:     time.Now(),
		}
		fRes, err := route.Filters.ExecuteRequest(fCtx)
		if err != nil {
			return nil, err
		}
		if fRes != nil && fRes.Action == ActionAbort {
			return nil, fmt.Errorf("request aborted by filter (status %d): %s", fRes.StatusCode, fRes.ErrorMessage)
		}
		if fRes != nil && fRes.Action == ActionTransform {
			for k, v := range fRes.Headers {
				req.Header[k] = v
			}
		}
	}

	// 3. Fault Injection Simulation
	if route != nil && route.FaultInjection != nil {
		if route.FaultInjection.DelayDuration > 0 && route.FaultInjection.DelayPercentage > 0 {
			if rand.Float64()*100 <= route.FaultInjection.DelayPercentage {
				time.Sleep(route.FaultInjection.DelayDuration)
			}
		}
		if route.FaultInjection.AbortStatusCode > 0 && route.FaultInjection.AbortPercentage > 0 {
			if rand.Float64()*100 <= route.FaultInjection.AbortPercentage {
				return nil, fmt.Errorf("fault injection aborted with status %d", route.FaultInjection.AbortStatusCode)
			}
		}
	}

	// 4. Destination Selection with Canary Header Overrides
	target := ""
	if route != nil && req != nil && len(route.CanaryRules) > 0 {
		for _, cr := range route.CanaryRules {
			if req.Header.Get(cr.HeaderName) == cr.HeaderValue {
				target = cr.TargetBackend
				break
			}
		}
	}
	if target == "" {
		var err error
		target, err = p.SelectDestination(serviceRef)
		if err != nil {
			return nil, err
		}
	}

	targetURL, err := url.Parse("http://" + target + req.URL.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid destination url: %w", err)
	}

	outReq, err := http.NewRequest(req.Method, targetURL.String(), req.Body)
	if err != nil {
		return nil, err
	}
	outReq.Header = req.Header.Clone()

	// Inject sidecarless mesh identity metadata headers
	if route != nil && route.MTLSMode != "" {
		outReq.Header.Set("X-StraitGateway-Mesh-Mode", string(route.MTLSMode))
	}
	outReq.Header.Set("X-Forwarded-Proto", "https")

	maxAttempts := 1
	if route != nil && route.Retries != nil && route.Retries.Attempts > 1 {
		maxAttempts = int(route.Retries.Attempts)
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		resp, err := p.httpClient.Do(outReq)
		if err == nil && resp.StatusCode < 500 {
			if route != nil && route.CircuitBreaker != nil {
				route.CircuitBreaker.RecordSuccess()
			}
			if route != nil && route.Filters != nil && fCtx != nil {
				_ = route.Filters.ExecuteResponse(resp, fCtx)
			}
			return resp, nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("upstream returned status %d", resp.StatusCode)
			_ = resp.Body.Close()
		}
		if route != nil && route.CircuitBreaker != nil {
			route.CircuitBreaker.RecordFailure()
		}
		time.Sleep(50 * time.Millisecond) // backoff
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxAttempts, lastErr)
}

// GetRoute returns an L7 route policy by serviceRef
func (p *NodeProxy) GetRoute(serviceRef string) (*ProxyRoute, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	r, exists := p.routes[serviceRef]
	return r, exists
}

// RemoveRoute deletes an L7 route policy
func (p *NodeProxy) RemoveRoute(serviceRef string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.routes, serviceRef)
}

// ListRoutes returns all registered L7 routes
func (p *NodeProxy) ListRoutes() []*ProxyRoute {
	p.mu.RLock()
	defer p.mu.RUnlock()

	list := make([]*ProxyRoute, 0, len(p.routes))
	for _, r := range p.routes {
		list = append(list, r)
	}
	return list
}

