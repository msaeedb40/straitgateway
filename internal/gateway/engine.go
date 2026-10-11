package gateway

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// BackendDestination represents a weighted upstream target
type BackendDestination struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Port      int32  `json:"port"`
	Weight    int32  `json:"weight"`
	TargetURL string `json:"targetUrl,omitempty"`
}

// HTTPRouteRuleConfig represents a compiled HTTPRoute rule
type HTTPRouteRuleConfig struct {
	Name            string
	Matches         []gatewayv1.HTTPRouteMatch
	BackendRefs     []BackendDestination
	TotalWeight     int32
	RequestHeaders  *gatewayv1.HTTPHeaderFilter
	ResponseHeaders *gatewayv1.HTTPHeaderFilter
	URLRewrite      *gatewayv1.HTTPURLRewriteFilter
}

// HTTPRouteEngine evaluates Gateway API HTTPRoute traffic rules
type HTTPRouteEngine struct {
	mu     sync.RWMutex
	routes map[string][]*HTTPRouteRuleConfig // Hostname -> Rules
	rng    *rand.Rand
}

// NewHTTPRouteEngine creates a new HTTP route engine
func NewHTTPRouteEngine() *HTTPRouteEngine {
	return &HTTPRouteEngine{
		routes: make(map[string][]*HTTPRouteRuleConfig),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// AddRoute registers rules for a specified hostname (or wildcard "*")
func (e *HTTPRouteEngine) AddRoute(hostname string, rules []*HTTPRouteRuleConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if hostname == "" {
		hostname = "*"
	}
	e.routes[strings.ToLower(hostname)] = rules
}

// RemoveRoute clears rules for a hostname
func (e *HTTPRouteEngine) RemoveRoute(hostname string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.routes, strings.ToLower(hostname))
}

// RouteResult holds the resolution outcome
type RouteResult struct {
	MatchedRule    string
	SelectedTarget *BackendDestination
	RewrittenPath  string
	ModifiedHeader http.Header
}

// Match evaluates an incoming HTTP request against compiled Gateway API rules
func (e *HTTPRouteEngine) Match(req *http.Request) (*RouteResult, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	host := strings.ToLower(req.Host)
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Try exact hostname match, then wildcard
	rules, exists := e.routes[host]
	if !exists {
		rules = e.routes["*"]
	}
	if len(rules) == 0 {
		return nil, false
	}

	for _, rule := range rules {
		if e.matchesRule(req, rule.Matches) {
			res := &RouteResult{
				MatchedRule:    rule.Name,
				SelectedTarget: e.selectBackend(rule),
				ModifiedHeader: make(http.Header),
			}

			// Apply Header Filters
			if rule.RequestHeaders != nil {
				for _, set := range rule.RequestHeaders.Set {
					req.Header.Set(string(set.Name), set.Value)
				}
				for _, add := range rule.RequestHeaders.Add {
					req.Header.Add(string(add.Name), add.Value)
				}
				for _, rem := range rule.RequestHeaders.Remove {
					req.Header.Del(rem)
				}
			}

			// Apply URL Rewrite
			if rule.URLRewrite != nil && rule.URLRewrite.Path != nil {
				if rule.URLRewrite.Path.ReplacePrefixMatch != nil {
					newPrefix := *rule.URLRewrite.Path.ReplacePrefixMatch
					matchedPrefix := ""
					for _, m := range rule.Matches {
						if m.Path != nil && m.Path.Value != nil && strings.HasPrefix(req.URL.Path, *m.Path.Value) {
							matchedPrefix = *m.Path.Value
							break
						}
					}
					remainder := req.URL.Path
					if matchedPrefix != "" {
						remainder = strings.TrimPrefix(remainder, matchedPrefix)
					}
					res.RewrittenPath = strings.TrimSuffix(newPrefix, "/") + "/" + strings.TrimPrefix(remainder, "/")
				}
			}

			return res, true
		}
	}

	return nil, false
}

func (e *HTTPRouteEngine) matchesRule(req *http.Request, matches []gatewayv1.HTTPRouteMatch) bool {
	if len(matches) == 0 {
		return true // Default match all if no conditions specified
	}

	for _, m := range matches {
		if e.matchesSingle(req, m) {
			return true
		}
	}
	return false
}

func (e *HTTPRouteEngine) matchesSingle(req *http.Request, m gatewayv1.HTTPRouteMatch) bool {
	// 1. Path Match
	if m.Path != nil {
		pathVal := "/"
		if m.Path.Value != nil {
			pathVal = *m.Path.Value
		}

		pathType := gatewayv1.PathMatchPathPrefix
		if m.Path.Type != nil {
			pathType = *m.Path.Type
		}

		switch pathType {
		case gatewayv1.PathMatchExact:
			if req.URL.Path != pathVal {
				return false
			}
		case gatewayv1.PathMatchPathPrefix:
			if !strings.HasPrefix(req.URL.Path, pathVal) {
				return false
			}
		case gatewayv1.PathMatchRegularExpression:
			matched, err := regexp.MatchString(pathVal, req.URL.Path)
			if err != nil || !matched {
				return false
			}
		}
	}

	// 2. HTTP Method Match
	if m.Method != nil {
		if req.Method != string(*m.Method) {
			return false
		}
	}

	// 3. Header Matches
	for _, h := range m.Headers {
		val := req.Header.Get(string(h.Name))
		hType := gatewayv1.HeaderMatchExact
		if h.Type != nil {
			hType = *h.Type
		}

		switch hType {
		case gatewayv1.HeaderMatchExact:
			if val != h.Value {
				return false
			}
		case gatewayv1.HeaderMatchRegularExpression:
			matched, err := regexp.MatchString(h.Value, val)
			if err != nil || !matched {
				return false
			}
		}
	}

	// 4. Query Param Matches
	for _, qp := range m.QueryParams {
		val := req.URL.Query().Get(string(qp.Name))
		qpType := gatewayv1.QueryParamMatchExact
		if qp.Type != nil {
			qpType = *qp.Type
		}

		switch qpType {
		case gatewayv1.QueryParamMatchExact:
			if val != qp.Value {
				return false
			}
		case gatewayv1.QueryParamMatchRegularExpression:
			matched, err := regexp.MatchString(qp.Value, val)
			if err != nil || !matched {
				return false
			}
		}
	}

	return true
}

// selectBackend performs weighted selection across backend targets
func (e *HTTPRouteEngine) selectBackend(rule *HTTPRouteRuleConfig) *BackendDestination {
	if len(rule.BackendRefs) == 0 {
		return nil
	}
	if len(rule.BackendRefs) == 1 || rule.TotalWeight <= 0 {
		return &rule.BackendRefs[0]
	}

	r := e.rng.Int31n(rule.TotalWeight)
	cur := int32(0)
	for i := range rule.BackendRefs {
		cur += rule.BackendRefs[i].Weight
		if r < cur {
			return &rule.BackendRefs[i]
		}
	}

	return &rule.BackendRefs[0]
}

// GRPCRouteEngine matches and routes gRPC calls
type GRPCRouteEngine struct {
	mu     sync.RWMutex
	routes map[string]*BackendDestination // Key: /service/method
}

// NewGRPCRouteEngine creates a gRPC route engine
func NewGRPCRouteEngine() *GRPCRouteEngine {
	return &GRPCRouteEngine{
		routes: make(map[string]*BackendDestination),
	}
}

// RegisterMethod adds a gRPC route
func (g *GRPCRouteEngine) RegisterMethod(service, method string, target *BackendDestination) {
	g.mu.Lock()
	defer g.mu.Unlock()
	key := fmt.Sprintf("/%s/%s", strings.TrimPrefix(service, "/"), method)
	g.routes[key] = target
}

// MatchMethod finds target backend for a gRPC request path
func (g *GRPCRouteEngine) MatchMethod(path string) (*BackendDestination, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	tgt, ok := g.routes[path]
	return tgt, ok
}
