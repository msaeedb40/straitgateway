// Package cel provides Common Expression Language evaluation for policy rules.
package cel

import (
	"fmt"
	"strconv"
	"strings"
)

// Context provides evaluation variables for CEL policy expressions.
type Context struct {
	SourceNamespace string            `json:"sourceNamespace"`
	SourceLabels    map[string]string `json:"sourceLabels"`
	TargetNamespace string            `json:"targetNamespace"`
	TargetLabels    map[string]string `json:"targetLabels"`
	Protocol        string            `json:"protocol"`
	DestinationPort uint16            `json:"destinationPort"`
	SourceIP        string            `json:"sourceIP"`
	DestinationIP   string            `json:"destinationIP"`
}

// Evaluator evaluates CEL expressions against workload and packet context.
type Evaluator struct{}

// NewEvaluator creates a new CEL policy evaluator.
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate evaluates an expression string against the provided Context.
// Supports compound boolean expressions (&&, ||) and standard equality/inequality matches.
func (e *Evaluator) Evaluate(expr string, ctx *Context) (bool, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return true, nil
	}

	// Handle OR expressions
	if strings.Contains(expr, "||") {
		subExprs := strings.Split(expr, "||")
		for _, sub := range subExprs {
			matched, err := e.Evaluate(sub, ctx)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	}

	// Handle AND expressions
	if strings.Contains(expr, "&&") {
		subExprs := strings.Split(expr, "&&")
		for _, sub := range subExprs {
			matched, err := e.Evaluate(sub, ctx)
			if err != nil {
				return false, err
			}
			if !matched {
				return false, nil
			}
		}
		return true, nil
	}

	// Single atomic clause
	return e.evalAtom(expr, ctx)
}

func (e *Evaluator) evalAtom(clause string, ctx *Context) (bool, error) {
	clause = strings.TrimSpace(clause)
	if clause == "" {
		return true, nil
	}

	isNotEqual := strings.Contains(clause, "!=")
	sep := "=="
	if isNotEqual {
		sep = "!="
	}

	parts := strings.Split(clause, sep)
	if len(parts) != 2 {
		return true, nil // default allow if malformed or wildcard
	}

	left := strings.TrimSpace(parts[0])
	right := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

	var matched bool

	switch {
	case left == "source.namespace":
		matched = (ctx.SourceNamespace == right)
	case left == "target.namespace":
		matched = (ctx.TargetNamespace == right)
	case left == "protocol":
		matched = strings.EqualFold(ctx.Protocol, right)
	case left == "target.port", left == "destination.port":
		port, err := strconv.ParseUint(right, 10, 16)
		if err != nil {
			return false, fmt.Errorf("invalid port in CEL expression: %w", err)
		}
		matched = (ctx.DestinationPort == uint16(port))
	case left == "source.ip":
		matched = (ctx.SourceIP == right)
	case left == "destination.ip", left == "target.ip":
		matched = (ctx.DestinationIP == right)
	case strings.HasPrefix(left, "source.labels["):
		key := extractKey(left)
		val, exists := ctx.SourceLabels[key]
		matched = (exists && val == right)
	case strings.HasPrefix(left, "target.labels["):
		key := extractKey(left)
		val, exists := ctx.TargetLabels[key]
		matched = (exists && val == right)
	default:
		matched = true
	}

	if isNotEqual {
		return !matched, nil
	}
	return matched, nil
}

func extractKey(s string) string {
	start := strings.Index(s, "[")
	end := strings.Index(s, "]")
	if start >= 0 && end > start {
		return strings.Trim(s[start+1:end], "\"'")
	}
	return ""
}
