package cel_test

import (
	"testing"

	"github.com/straitgateway/straitgateway/internal/policy/cel"
)

func TestCELEvaluator(t *testing.T) {
	eval := cel.NewEvaluator()
	ctx := &cel.Context{
		SourceNamespace: "prod",
		SourceLabels:    map[string]string{"app": "frontend", "env": "prod"},
		TargetNamespace: "backend",
		TargetLabels:    map[string]string{"app": "payment"},
		Protocol:        "TCP",
		DestinationPort: 443,
		SourceIP:        "10.244.1.10",
		DestinationIP:   "10.244.2.20",
	}

	tests := []struct {
		expr     string
		expected bool
	}{
		{"source.namespace == 'prod'", true},
		{"source.namespace == 'dev'", false},
		{"target.port == 443", true},
		{"target.port == 80", false},
		{"source.labels['app'] == 'frontend' && target.labels['app'] == 'payment'", true},
		{"source.labels['env'] == 'dev' || target.port == 443", true},
		{"protocol == 'TCP' && target.namespace == 'backend'", true},
		{"protocol == 'UDP'", false},
		{"source.namespace != 'dev'", true},
	}

	for _, tc := range tests {
		matched, err := eval.Evaluate(tc.expr, ctx)
		if err != nil {
			t.Fatalf("Evaluate(%q) error = %v", tc.expr, err)
		}
		if matched != tc.expected {
			t.Errorf("Evaluate(%q) = %v, want %v", tc.expr, matched, tc.expected)
		}
	}
}
