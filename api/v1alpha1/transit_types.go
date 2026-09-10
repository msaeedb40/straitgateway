// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TransitTopology defines the multi-cluster transit topology type.
// +kubebuilder:validation:Enum=HubAndSpoke;Mesh;PeerToPeer;GatewayToGateway
type TransitTopology string

const (
	TopologyHubAndSpoke       TransitTopology = "HubAndSpoke"
	TopologyMesh              TransitTopology = "Mesh"
	TopologyPeerToPeer        TransitTopology = "PeerToPeer"
	TopologyGatewayToGateway  TransitTopology = "GatewayToGateway"

	TransitTopologyHubSpoke      = TopologyHubAndSpoke
	TransitTopologyMesh          = TopologyMesh
	TransitTopologyPeerToPeer    = TopologyPeerToPeer
	TransitTopologyGatewayToGateway = TopologyGatewayToGateway
)

// TransitEncryption defines the overlay encryption mode.
// +kubebuilder:validation:Enum=None;WireGuard;IPsec
type TransitEncryption string

const (
	TransitEncryptionNone      TransitEncryption = "None"
	TransitEncryptionWireGuard TransitEncryption = "WireGuard"
	TransitEncryptionIPsec     TransitEncryption = "IPsec"
)

// TransitGatewaySpec defines the desired state of TransitGateway.
type TransitGatewaySpec struct {
	// ClusterID is the unique identifier for this cluster in the transit mesh.
	// +kubebuilder:validation:Required
	ClusterID string `json:"clusterID"`

	// Topology defines the multi-cluster transit networking topology.
	// +kubebuilder:default=Mesh
	Topology TransitTopology `json:"topology"`

	// Encryption defines the overlay encryption mode between transit peers.
	// +kubebuilder:default=WireGuard
	Encryption TransitEncryption `json:"encryption"`

	// BackboneSegmentID is the segment used for backbone inter-cluster routing (default: 0).
	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	BackboneSegmentID uint32 `json:"backboneSegmentID"`

	// NodeSelector restricts which nodes run as transit gateway nodes.
	// +optional
	NodeSelector *metav1.LabelSelector `json:"nodeSelector,omitempty"`
}

// TransitGatewayStatus defines the observed state of TransitGateway.
type TransitGatewayStatus struct {
	// ObservedGeneration is the last generation reconciled.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// ActivePeers is the number of currently active transit peers.
	// +optional
	ActivePeers int32 `json:"activePeers,omitempty"`

	// Conditions holds reconciliation status conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// TransitGateway is the Schema for multi-cluster transit gateway configuration.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=tgw,categories=straitgateway
// +kubebuilder:printcolumn:name="Topology",type=string,JSONPath=`.spec.topology`
// +kubebuilder:printcolumn:name="Encryption",type=string,JSONPath=`.spec.encryption`
// +kubebuilder:printcolumn:name="Peers",type=integer,JSONPath=`.status.activePeers`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type TransitGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransitGatewaySpec   `json:"spec,omitempty"`
	Status TransitGatewayStatus `json:"status,omitempty"`
}

// TransitGatewayList contains a list of TransitGateway.
// +kubebuilder:object:root=true
type TransitGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitGateway `json:"items"`
}

// TransitSegmentSpec defines the desired state of a TransitSegment.
type TransitSegmentSpec struct {
	// SegmentID is the 32-bit segment identifier. Segment 0 is the backbone.
	// All segments are isolated by default; connectivity is via backbone or explicit routes.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	SegmentID uint32 `json:"segmentID"`

	// Description is a human-readable description of the segment's purpose.
	// +optional
	Description string `json:"description,omitempty"`
}

// TransitSegmentStatus defines the observed state of a TransitSegment.
type TransitSegmentStatus struct {
	// ObservedGeneration is the last generation reconciled.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds reconciliation status conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// TransitSegment is the Schema for 32-bit network segment isolation.
// Segment 0 is the backbone segment; all others are isolated until explicitly attached.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=ts,categories=straitgateway
// +kubebuilder:printcolumn:name="SegmentID",type=integer,JSONPath=`.spec.segmentID`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type TransitSegment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransitSegmentSpec   `json:"spec,omitempty"`
	Status TransitSegmentStatus `json:"status,omitempty"`
}

// TransitSegmentList contains a list of TransitSegment.
// +kubebuilder:object:root=true
type TransitSegmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitSegment `json:"items"`
}

// TransitAttachmentRef references a gateway or cluster attachment point.
type TransitAttachmentRef struct {
	// Name is the name of the referenced attachment (TransitGateway or cluster name).
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// SegmentID is the segment this attachment belongs to.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	SegmentID uint32 `json:"segmentID"`
}

// TransitSegmentAttachmentSpec defines the desired state of a TransitSegmentAttachment.
type TransitSegmentAttachmentSpec struct {
	// Attachments lists the segment attachment points (gateways or clusters) to connect.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=2
	Attachments []TransitAttachmentRef `json:"attachments"`
}

// TransitSegmentAttachmentStatus defines the observed state.
type TransitSegmentAttachmentStatus struct {
	// ObservedGeneration is the last generation reconciled.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions holds reconciliation status conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// TransitSegmentAttachment connects clusters or gateways to one or more segments.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=tsa,categories=straitgateway
type TransitSegmentAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransitSegmentAttachmentSpec   `json:"spec,omitempty"`
	Status TransitSegmentAttachmentStatus `json:"status,omitempty"`
}

// TransitSegmentAttachmentList contains a list of TransitSegmentAttachment.
// +kubebuilder:object:root=true
type TransitSegmentAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitSegmentAttachment `json:"items"`
}

// TransitSegmentRouteSpec defines the desired state of a TransitSegmentRoute.
type TransitSegmentRouteSpec struct {
	// CIDR is the destination CIDR for this transit route (e.g. 0.0.0.0/0).
	// +kubebuilder:validation:Required
	CIDR string `json:"cidr"`

	// NextHop is the name of the TransitSegmentAttachment that is the next hop.
	// +kubebuilder:validation:Required
	NextHop string `json:"nextHop"`

	// SegmentID is the segment this route belongs to.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	SegmentID uint32 `json:"segmentID"`
}

// TransitSegmentRouteStatus defines the observed state.
type TransitSegmentRouteStatus struct {
	// ObservedGeneration is the last generation reconciled.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions holds reconciliation status conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// TransitSegmentRoute defines a CIDR route within a transit segment.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=tsr,categories=straitgateway
// +kubebuilder:printcolumn:name="CIDR",type=string,JSONPath=`.spec.cidr`
// +kubebuilder:printcolumn:name="NextHop",type=string,JSONPath=`.spec.nextHop`
// +kubebuilder:printcolumn:name="SegmentID",type=integer,JSONPath=`.spec.segmentID`
type TransitSegmentRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransitSegmentRouteSpec   `json:"spec,omitempty"`
	Status TransitSegmentRouteStatus `json:"status,omitempty"`
}

// TransitSegmentRouteList contains a list of TransitSegmentRoute.
// +kubebuilder:object:root=true
type TransitSegmentRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitSegmentRoute `json:"items"`
}
