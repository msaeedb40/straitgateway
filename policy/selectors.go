// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"sort"

	sgv1 "github.com/msaeedb40/straitgateway/api/v1alpha1"
	"github.com/msaeedb40/straitgateway/identity"
	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
)

// SelectorType indicates which entity type a selector targets.
type SelectorType string

const (
	SelectorNamespace SelectorType = "namespace"
	SelectorPod       SelectorType = "pod"
	SelectorCluster   SelectorType = "cluster"
	SelectorSegment   SelectorType = "segment"
	SelectorGateway   SelectorType = "gateway"
	SelectorRoute     SelectorType = "route"
)

// IdentityRange represents a range of identities matched by a selector.
type IdentityRange struct {
	Min sgtypes.Identity
	Max sgtypes.Identity
}

// ResolveSelectors takes a StraitNetworkPolicy's selectors and returns the
// set of identity ranges they match.
func ResolveSelectors(
	selectors []sgv1.StraitPolicySelector,
	alloc *identity.Allocator,
) []IdentityRange {
	var ranges []IdentityRange

	for _, sel := range selectors {
		if sel.PodSelector != nil && len(sel.PodSelector.MatchLabels) > 0 {
			// Find all identities whose labels are a superset of the selector.
			for _, entry := range alloc.AllEntries() {
				if labelsMatch(entry.Labels, sel.PodSelector.MatchLabels) {
					ranges = append(ranges, IdentityRange{
						Min: entry.Identity,
						Max: entry.Identity,
					})
				}
			}
		}
		if sel.NamespaceSelector != nil && len(sel.NamespaceSelector.MatchLabels) > 0 {
			for _, entry := range alloc.AllEntries() {
				if labelsMatch(entry.Labels, sel.NamespaceSelector.MatchLabels) {
					ranges = append(ranges, IdentityRange{
						Min: entry.Identity,
						Max: entry.Identity,
					})
				}
			}
		}
	}

	return mergeRanges(ranges)
}

// labelsMatch returns true if target contains all key=value pairs from selector.
func labelsMatch(target, selector map[string]string) bool {
	for k, v := range selector {
		if target[k] != v {
			return false
		}
	}
	return true
}

// mergeRanges sorts and merges contiguous identity ranges.
func mergeRanges(ranges []IdentityRange) []IdentityRange {
	if len(ranges) == 0 {
		return nil
	}
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].Min < ranges[j].Min })
	merged := []IdentityRange{ranges[0]}
	for _, r := range ranges[1:] {
		last := &merged[len(merged)-1]
		if r.Min <= last.Max+1 {
			if r.Max > last.Max {
				last.Max = r.Max
			}
		} else {
			merged = append(merged, r)
		}
	}
	return merged
}
