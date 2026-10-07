package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RouteTargetType represents next-hop destination target
type RouteTargetType string

const (
	TargetTypeNode      RouteTargetType = "Node"
	TargetTypePeer      RouteTargetType = "GatewayPeer"
	TargetTypeWireGuard RouteTargetType = "WireGuard"
	TargetTypeInterface RouteTargetType = "Interface"
	TargetTypeBlackhole RouteTargetType = "Blackhole"
)

// TransitRouteSpec defines the desired state of TransitRoute
type TransitRouteSpec struct {
	// DestinationCIDR is the IPv4/IPv6 target network prefix (e.g. "10.244.0.0/16")
	DestinationCIDR string `json:"destinationCidr"`

	// TargetType indicates the category of next-hop
	TargetType RouteTargetType `json:"targetType"`

	// TargetRef identifies the specific resource or node name
	// +optional
	TargetRef string `json:"targetRef,omitempty"`

	// NextHopIP is the IP address to forward matching traffic to
	// +optional
	NextHopIP string `json:"nextHopIp,omitempty"`

	// OutgoingInterface is the egress Linux device name (e.g. "wg0", "eth0")
	// +optional
	OutgoingInterface string `json:"outgoingInterface,omitempty"`

	// VRF is the routing table domain ID (default: 0)
	// +optional
	VRF uint32 `json:"vrf,omitempty"`

	// Metric specifies route preference / administrative distance (lower is preferred)
	// +optional
	Metric uint32 `json:"metric,omitempty"`

	// Communities are optional BGP community strings attached to this route
	// +optional
	Communities []string `json:"communities,omitempty"`
}

// TransitRouteStatus defines the observed state of TransitRoute
type TransitRouteStatus struct {
	// Programmed indicates whether this route is actively synced into eBPF route maps
	Programmed bool `json:"programmed"`

	// NodesProgrammed lists the nodes where this route has been installed into BPF maps
	// +optional
	NodesProgrammed []string `json:"nodesProgrammed,omitempty"`

	// LastSyncTime is the timestamp of the last successful map update
	// +optional
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`

	// Message provides human-readable error or sync details
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="CIDR",type="string",JSONPath=".spec.destinationCidr"
// +kubebuilder:printcolumn:name="TargetType",type="string",JSONPath=".spec.targetType"
// +kubebuilder:printcolumn:name="NextHop",type="string",JSONPath=".spec.nextHopIp"
// +kubebuilder:printcolumn:name="VRF",type="integer",JSONPath=".spec.vrf"
// +kubebuilder:printcolumn:name="Programmed",type="boolean",JSONPath=".status.programmed"

// TransitRoute is the Schema for the transitroutes API
type TransitRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransitRouteSpec   `json:"spec,omitempty"`
	Status TransitRouteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TransitRouteList contains a list of TransitRoute
type TransitRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitRoute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TransitRoute{}, &TransitRouteList{})
}
