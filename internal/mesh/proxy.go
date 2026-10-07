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

// ProxyRoute holds L7 routing and traffic split rules
type ProxyRoute struct {
	ServiceRef   string
	TrafficSplit []v1alpha1.TrafficSplitTarget
	Retries      *v1alpha1.RetryPolicy
	Timeout      time.Duration
	MTLSMode     v1alpha1.MTLSMode
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

// ExecuteRequest forwards a request with retries and timeout
func (p *NodeProxy) ExecuteRequest(req *http.Request, serviceRef string) (*http.Response, error) {
	target, err := p.SelectDestination(serviceRef)
	if err != nil {
		return nil, err
	}

	p.mu.RLock()
	route := p.routes[serviceRef]
	p.mu.RUnlock()

	targetURL, err := url.Parse("http://" + target + req.URL.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid destination url: %w", err)
	}

	outReq, err := http.NewRequest(req.Method, targetURL.String(), req.Body)
	if err != nil {
		return nil, err
	}
	outReq.Header = req.Header.Clone()

	maxAttempts := 1
	if route != nil && route.Retries != nil && route.Retries.Attempts > 1 {
		maxAttempts = int(route.Retries.Attempts)
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		resp, err := p.httpClient.Do(outReq)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("upstream returned status %d", resp.StatusCode)
			_ = resp.Body.Close()
		}
		time.Sleep(50 * time.Millisecond) // backoff
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxAttempts, lastErr)
}
