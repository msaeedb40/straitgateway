// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// ReferenceGrantChecker validates cross-namespace references per Gateway API spec.
type ReferenceGrantChecker struct {
	client client.Client
	log    *zap.Logger
}

// NewReferenceGrantChecker creates a new checker.
func NewReferenceGrantChecker(c client.Client, log *zap.Logger) *ReferenceGrantChecker {
	return &ReferenceGrantChecker{client: c, log: log}
}

// IsAllowed checks if a cross-namespace reference is permitted by a ReferenceGrant.
func (r *ReferenceGrantChecker) IsAllowed(ctx context.Context, fromNS, fromKind, toNS, toName, toKind string) (bool, error) {
	var grants gwapiv1beta1.ReferenceGrantList
	if err := r.client.List(ctx, &grants, client.InNamespace(toNS)); err != nil {
		return false, fmt.Errorf("listing ReferenceGrants in %s: %w", toNS, err)
	}

	for _, grant := range grants.Items {
		for _, from := range grant.Spec.From {
			if string(from.Namespace) == fromNS && string(from.Kind) == fromKind {
				for _, to := range grant.Spec.To {
					if string(to.Kind) == toKind {
						if to.Name == nil || string(*to.Name) == toName {
							return true, nil
						}
					}
				}
			}
		}
	}

	return false, nil
}
