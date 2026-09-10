// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package gateway implements the Gateway API controller for straitgateway.
// GatewayClass controller name: straitgateway.io/skgateway
//
// Handled resources: GatewayClass, Gateway, HTTPRoute, GRPCRoute, TLSRoute, TCPRoute, UDPRoute, ReferenceGrant
package gateway

const (
	// ControllerName is the Gateway API GatewayClass controller name for straitgateway.
	ControllerName = "straitgateway.io/skgateway"

	// GatewayClassName is the name of the managed GatewayClass resource.
	GatewayClassName = "skgateway"
)
