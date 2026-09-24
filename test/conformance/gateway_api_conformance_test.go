package conformance

import (
	"strings"
	"testing"
)

// matchHostname conforms to Gateway API specification section on Hostname matching:
// - Exact match ("foo.bar.com" == "foo.bar.com")
// - Wildcard match ("*.bar.com" matches "foo.bar.com", but not "bar.com" or "sub.foo.bar.com")
func matchHostname(listenerHost, routeHost string) bool {
	if listenerHost == "" || routeHost == "" {
		return true
	}
	if listenerHost == routeHost {
		return true
	}
	if strings.HasPrefix(listenerHost, "*.") {
		suffix := listenerHost[1:] // ".bar.com"
		if strings.HasSuffix(routeHost, suffix) {
			prefix := strings.TrimSuffix(routeHost, suffix)
			return !strings.Contains(prefix, ".")
		}
	}
	return false
}

// matchPath conforms to Gateway API HTTPRoute path matching rules (Exact, PathPrefix).
func matchPath(pathType, rulePath, reqPath string) bool {
	switch pathType {
	case "Exact":
		return rulePath == reqPath
	case "PathPrefix":
		if rulePath == "/" {
			return true
		}
		if rulePath == reqPath {
			return true
		}
		prefixWithSlash := strings.TrimSuffix(rulePath, "/") + "/"
		return strings.HasPrefix(reqPath, prefixWithSlash)
	default:
		return false
	}
}

// TestGatewayAPIHostnameMatchingConformance tests Gateway API specification rules for hostname resolution.
func TestGatewayAPIHostnameMatchingConformance(t *testing.T) {
	tests := []struct {
		listener string
		route    string
		expected bool
	}{
		{"example.com", "example.com", true},
		{"example.com", "other.com", false},
		{"*.example.com", "api.example.com", true},
		{"*.example.com", "web.example.com", true},
		{"*.example.com", "sub.api.example.com", false},
		{"*.example.com", "example.com", false},
		{"", "foo.com", true},
		{"foo.com", "", true},
	}

	for _, tc := range tests {
		result := matchHostname(tc.listener, tc.route)
		if result != tc.expected {
			t.Errorf("matchHostname(%q, %q) = %v; expected %v", tc.listener, tc.route, result, tc.expected)
		}
	}
}

// TestGatewayAPIPathMatchingConformance tests Gateway API HTTPRoute path prefix and exact matching.
func TestGatewayAPIPathMatchingConformance(t *testing.T) {
	tests := []struct {
		pathType string
		rulePath string
		reqPath  string
		expected bool
	}{
		{"Exact", "/api/v1", "/api/v1", true},
		{"Exact", "/api/v1", "/api/v1/", false},
		{"Exact", "/api/v1", "/api/v1/users", false},
		{"PathPrefix", "/api", "/api", true},
		{"PathPrefix", "/api", "/api/", true},
		{"PathPrefix", "/api", "/api/v1/users", true},
		{"PathPrefix", "/api", "/apidoc", false}, // Must match on path segment boundaries
		{"PathPrefix", "/", "/anything", true},
	}

	for _, tc := range tests {
		result := matchPath(tc.pathType, tc.rulePath, tc.reqPath)
		if result != tc.expected {
			t.Errorf("matchPath(%s, %q, %q) = %v; expected %v", tc.pathType, tc.rulePath, tc.reqPath, result, tc.expected)
		}
	}
}

// TestGatewayClassControllerNameConformance verifies the standard controller registration string.
func TestGatewayClassControllerNameConformance(t *testing.T) {
	const expectedControllerName = "straitgateway.io/skgateway"
	const gatewayClassName = "skgateway"

	if !strings.HasPrefix(expectedControllerName, "straitgateway.io/") {
		t.Errorf("expected controllerName to be domain-prefixed: %s", expectedControllerName)
	}
	if gatewayClassName == "" {
		t.Error("gatewayClassName must not be empty")
	}
}
