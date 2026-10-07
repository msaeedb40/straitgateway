package mesh

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// FilterAction indicates how the proxy should proceed after filter execution
type FilterAction string

const (
	ActionContinue  FilterAction = "continue"
	ActionAbort     FilterAction = "abort"
	ActionTransform FilterAction = "transform"
)

// FilterContext carries request metadata across the filter pipeline
type FilterContext struct {
	Request         *http.Request
	Headers         http.Header
	Attributes      map[string]interface{}
	ClientIdentity  uint32
	TargetService   string
	Timestamp       time.Time
}

// FilterResult holds the outcome of filter execution
type FilterResult struct {
	Action       FilterAction
	StatusCode   int
	ErrorMessage string
	Headers      http.Header
}

// L7Filter defines the interface for sidecarless mesh L7 extension filters
type L7Filter interface {
	Name() string
	ProcessRequest(ctx *FilterContext) (*FilterResult, error)
	ProcessResponse(resp *http.Response, ctx *FilterContext) error
}

// FilterPipeline executes a sequence of L7 filters in order
type FilterPipeline struct {
	mu      sync.RWMutex
	filters []L7Filter
}

// NewFilterPipeline creates an empty FilterPipeline
func NewFilterPipeline() *FilterPipeline {
	return &FilterPipeline{
		filters: make([]L7Filter, 0),
	}
}

// AddFilter registers an L7 filter into the pipeline
func (p *FilterPipeline) AddFilter(filter L7Filter) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.filters = append(p.filters, filter)
}

// ExecuteRequest runs all registered request filters
func (p *FilterPipeline) ExecuteRequest(ctx *FilterContext) (*FilterResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, f := range p.filters {
		res, err := f.ProcessRequest(ctx)
		if err != nil {
			return &FilterResult{
				Action:       ActionAbort,
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: fmt.Sprintf("filter %s error: %v", f.Name(), err),
			}, err
		}
		if res != nil && res.Action == ActionAbort {
			return res, nil
		}
		if res != nil && res.Action == ActionTransform {
			for k, v := range res.Headers {
				ctx.Headers[k] = v
			}
		}
	}

	return &FilterResult{Action: ActionContinue}, nil
}

// ExecuteResponse runs all registered response filters
func (p *FilterPipeline) ExecuteResponse(resp *http.Response, ctx *FilterContext) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, f := range p.filters {
		if err := f.ProcessResponse(resp, ctx); err != nil {
			return fmt.Errorf("response filter %s error: %w", f.Name(), err)
		}
	}
	return nil
}

// -------------------------------------------------------------
// Built-in L7 Filters
// -------------------------------------------------------------

// HeaderTransformFilter modifies request and response headers
type HeaderTransformFilter struct {
	NameTag       string
	AddHeaders    map[string]string
	RemoveHeaders []string
}

func (h *HeaderTransformFilter) Name() string {
	if h.NameTag != "" {
		return h.NameTag
	}
	return "header-transform"
}

func (h *HeaderTransformFilter) ProcessRequest(ctx *FilterContext) (*FilterResult, error) {
	headers := make(http.Header)
	for k, v := range h.AddHeaders {
		headers.Set(k, v)
		ctx.Headers.Set(k, v)
	}
	for _, k := range h.RemoveHeaders {
		ctx.Headers.Del(k)
	}
	return &FilterResult{Action: ActionTransform, Headers: headers}, nil
}

func (h *HeaderTransformFilter) ProcessResponse(resp *http.Response, ctx *FilterContext) error {
	if resp != nil {
		for k, v := range h.AddHeaders {
			resp.Header.Set("X-StraitGateway-"+k, v)
		}
	}
	return nil
}

// RateLimitFilter limits requests per second based on identity or client IP
type RateLimitFilter struct {
	mu        sync.Mutex
	LimitRPS  int
	tokens    map[string]int
	lastCheck time.Time
}

// NewRateLimitFilter creates a token bucket rate limiter
func NewRateLimitFilter(rps int) *RateLimitFilter {
	return &RateLimitFilter{
		LimitRPS:  rps,
		tokens:    make(map[string]int),
		lastCheck: time.Now(),
	}
}

func (r *RateLimitFilter) Name() string {
	return "rate-limiter"
}

func (r *RateLimitFilter) ProcessRequest(ctx *FilterContext) (*FilterResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if now.Sub(r.lastCheck) >= time.Second {
		r.tokens = make(map[string]int)
		r.lastCheck = now
	}

	key := fmt.Sprintf("id:%d", ctx.ClientIdentity)
	if ctx.ClientIdentity == 0 && ctx.Request != nil {
		key = ctx.Request.RemoteAddr
	}

	current := r.tokens[key]
	if current >= r.LimitRPS {
		return &FilterResult{
			Action:       ActionAbort,
			StatusCode:   http.StatusTooManyRequests,
			ErrorMessage: "rate limit exceeded",
		}, nil
	}

	r.tokens[key] = current + 1
	return &FilterResult{Action: ActionContinue}, nil
}

func (r *RateLimitFilter) ProcessResponse(resp *http.Response, ctx *FilterContext) error {
	return nil
}

// JWTAuthFilter inspects authorization bearer tokens
type JWTAuthFilter struct {
	ExpectedPrefix string
}

func (j *JWTAuthFilter) Name() string {
	return "jwt-auth"
}

func (j *JWTAuthFilter) ProcessRequest(ctx *FilterContext) (*FilterResult, error) {
	authHeader := ctx.Headers.Get("Authorization")
	if authHeader == "" {
		return &FilterResult{
			Action:       ActionAbort,
			StatusCode:   http.StatusUnauthorized,
			ErrorMessage: "missing authorization header",
		}, nil
	}

	prefix := j.ExpectedPrefix
	if prefix == "" {
		prefix = "Bearer "
	}

	if !strings.HasPrefix(authHeader, prefix) {
		return &FilterResult{
			Action:       ActionAbort,
			StatusCode:   http.StatusForbidden,
			ErrorMessage: "invalid authorization format",
		}, nil
	}

	token := strings.TrimPrefix(authHeader, prefix)
	if len(token) < 8 {
		return &FilterResult{
			Action:       ActionAbort,
			StatusCode:   http.StatusUnauthorized,
			ErrorMessage: "malformed token payload",
		}, nil
	}

	// Successful authentication
	ctx.Attributes["auth_token"] = token
	return &FilterResult{Action: ActionContinue}, nil
}

func (j *JWTAuthFilter) ProcessResponse(resp *http.Response, ctx *FilterContext) error {
	return nil
}
