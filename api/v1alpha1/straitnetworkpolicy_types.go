// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PolicyAction defines the action to take on matched traffic.
// +kubebuilder:validation:Enum=Allow;Deny;Reject
type PolicyAction string

const (
	PolicyActionAllow  PolicyAction = "Allow"
	PolicyActionDeny   PolicyAction = "Deny"
	PolicyActionReject PolicyAction = "Reject"
)

// PolicyDirection defines traffic direction.
// +kubebuilder:validation:Enum=Ingress;Egress
type PolicyDirection string

const (
	PolicyDirectionIngress PolicyDirection = "Ingress"
	PolicyDirectionEgress  PolicyDirection = "Egress"
)

// PolicyProtocol defines the network protocol for a policy rule.
// +kubebuilder:validation:Enum=TCP;UDP;ICMP;SCTP;Any
type PolicyProtocol string

const (
	PolicyProtocolTCP  PolicyProtocol = "TCP"
	PolicyProtocolUDP  PolicyProtocol = "UDP"
	PolicyProtocolICMP PolicyProtocol = "ICMP"
	PolicyProtocolSCTP PolicyProtocol = "SCTP"
	PolicyProtocolAny  PolicyProtocol = "Any"
)

// StraitPolicySelector defines a multi-dimensional traffic endpoint selector.
type StraitPolicySelector struct {
	// NamespaceSelector selects namespaces by label.
	// +optional
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`

	// PodSelector selects pods by label within the namespace.
	// +optional
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty"`

	// ClusterSelector selects remote clusters by label.
	// +optional
	ClusterSelector *metav1.LabelSelector `json:"clusterSelector,omitempty"`

	// SegmentSelector selects transit segments by label.
	// +optional
	SegmentSelector *metav1.LabelSelector `json:"segmentSelector,omitempty"`

	// GatewaySelector selects Gateway API Gateway resources.
	// +optional
	GatewaySelector *metav1.LabelSelector `json:"gatewaySelector,omitempty"`

	// HTTPRouteSelector selects Gateway API HTTPRoute resources.
	// +optional
	HTTPRouteSelector *metav1.LabelSelector `json:"httprouteSelector,omitempty"`

	// GRPCRouteSelector selects Gateway API GRPCRoute resources.
	// +optional
	GRPCRouteSelector *metav1.LabelSelector `json:"grpcrouteSelector,omitempty"`

	// TLSRouteSelector selects Gateway API TLSRoute resources.
	// +optional
	TLSRouteSelector *metav1.LabelSelector `json:"tlsrouteSelector,omitempty"`

	// TCPRouteSelector selects Gateway API TCPRoute resources.
	// +optional
	TCPRouteSelector *metav1.LabelSelector `json:"tcprouteSelector,omitempty"`

	// UDPRouteSelector selects Gateway API UDPRoute resources.
	// +optional
	UDPRouteSelector *metav1.LabelSelector `json:"udprouteSelector,omitempty"`
}

// PolicyPort defines a port and protocol for a policy rule.
type PolicyPort struct {
	// Protocol for this port. Defaults to TCP.
	// +optional
	// +kubebuilder:default=TCP
	Protocol PolicyProtocol `json:"protocol,omitempty"`

	// Port number. If omitted, all ports match.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port *int32 `json:"port,omitempty"`

	// EndPort defines the end of a port range (inclusive). Must be >= Port.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	EndPort *int32 `json:"endPort,omitempty"`
}

// PolicyRule defines a single traffic rule with selectors, ports, direction, and action.
type PolicyRule struct {
	// Direction specifies Ingress or Egress.
	// +kubebuilder:validation:Required
	Direction PolicyDirection `json:"direction"`

	// From lists selectors for ingress traffic sources.
	// +optional
	From []StraitPolicySelector `json:"from,omitempty"`

	// To lists selectors for egress traffic destinations.
	// +optional
	To []StraitPolicySelector `json:"to,omitempty"`

	// Ports restricts the rule to specific ports/protocols.
	// +optional
	Ports []PolicyPort `json:"ports,omitempty"`

	// Action defines what to do with matched traffic.
	// +kubebuilder:validation:Required
	Action PolicyAction `json:"action"`
}

// StraitNetworkPolicySpec defines the desired state of StraitNetworkPolicy.
type StraitNetworkPolicySpec struct {
	// Priority determines rule evaluation order. Lower value = higher priority.
	// Range: 0–255. Default: 100.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=255
	// +kubebuilder:default=100
	Priority int32 `json:"priority"`

	// PodSelector selects the pods this policy applies to within the namespace.
	// An empty selector matches all pods in the namespace.
	// +kubebuilder:validation:Required
	PodSelector metav1.LabelSelector `json:"podSelector"`

	// Rules defines the ordered list of policy rules.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Rules []PolicyRule `json:"rules"`
}

// StraitNetworkPolicyStatus defines the observed state of StraitNetworkPolicy.
type StraitNetworkPolicyStatus struct {
	// ObservedGeneration is the last generation reconciled by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// CompiledRules is the number of eBPF rules compiled from this policy.
	// +optional
	CompiledRules int32 `json:"compiledRules,omitempty"`

	// Conditions holds policy reconciliation conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// StraitNetworkPolicy is the Schema for straitgateway extended NetworkPolicy.
// It supports multi-dimensional selectors (namespace, pod, cluster, segment, gateway, routes)
// and explicit Allow/Deny/Reject actions with priority-based evaluation.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=snp,categories=straitgateway
// +kubebuilder:printcolumn:name="Priority",type=integer,JSONPath=`.spec.priority`
// +kubebuilder:printcolumn:name="Rules",type=integer,JSONPath=`.status.compiledRules`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type StraitNetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StraitNetworkPolicySpec   `json:"spec,omitempty"`
	Status StraitNetworkPolicyStatus `json:"status,omitempty"`
}

// StraitNetworkPolicyList contains a list of StraitNetworkPolicy.
// +kubebuilder:object:root=true
type StraitNetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StraitNetworkPolicy `json:"items"`
}
