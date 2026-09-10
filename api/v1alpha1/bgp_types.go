// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BGPPeerSpec defines the desired state of a BGPPeer.
type BGPPeerSpec struct {
	// PeerAddress is the IP address of the BGP peer.
	// +kubebuilder:validation:Required
	PeerAddress string `json:"peerAddress"`

	// PeerASN is the Autonomous System Number of the BGP peer.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=4294967295
	PeerASN uint32 `json:"peerASN"`

	// LocalASN is the local Autonomous System Number.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=4294967295
	LocalASN uint32 `json:"localASN"`

	// HoldTime is the BGP hold time in seconds. Default: 90.
	// +kubebuilder:default=90
	// +kubebuilder:validation:Minimum=3
	HoldTime int32 `json:"holdTime"`

	// KeepaliveTime is the BGP keepalive interval in seconds. Default: 30.
	// +kubebuilder:default=30
	// +kubebuilder:validation:Minimum=1
	KeepaliveTime int32 `json:"keepaliveTime"`

	// Password is the optional MD5 TCP authentication password.
	// +optional
	Password string `json:"password,omitempty"`

	// BFDEnabled enables BFD (Bidirectional Forwarding Detection) for fast failure detection.
	// +kubebuilder:default=false
	BFDEnabled bool `json:"bfdEnabled"`

	// ImportPolicy is the name of the route import filter policy.
	// +optional
	ImportPolicy string `json:"importPolicy,omitempty"`

	// ExportPolicy is the name of the route export filter policy.
	// +optional
	ExportPolicy string `json:"exportPolicy,omitempty"`

	// Prefixes is the list of CIDRs to advertise to this peer.
	// +optional
	Prefixes []string `json:"prefixes,omitempty"`
}

// BGPPeerStatus defines the observed state of a BGPPeer.
type BGPPeerStatus struct {
	// SessionState is the current BGP session state (Idle/Connect/Active/OpenSent/OpenConfirm/Established).
	// +optional
	SessionState string `json:"sessionState,omitempty"`

	// PrefixesAdvertised is the number of prefixes currently advertised to this peer.
	// +optional
	PrefixesAdvertised int32 `json:"prefixesAdvertised,omitempty"`

	// PrefixesReceived is the number of prefixes received from this peer.
	// +optional
	PrefixesReceived int32 `json:"prefixesReceived,omitempty"`

	// ObservedGeneration is the last generation reconciled.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds reconciliation status conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// BGPPeer is the Schema for a BGP peering relationship.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=bgpp,categories=straitgateway
// +kubebuilder:printcolumn:name="PeerAddress",type=string,JSONPath=`.spec.peerAddress`
// +kubebuilder:printcolumn:name="PeerASN",type=integer,JSONPath=`.spec.peerASN`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.sessionState`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type BGPPeer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BGPPeerSpec   `json:"spec,omitempty"`
	Status BGPPeerStatus `json:"status,omitempty"`
}

// BGPPeerList contains a list of BGPPeer.
// +kubebuilder:object:root=true
type BGPPeerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BGPPeer `json:"items"`
}

// BFDSessionSpec defines the desired state of a BFD session.
type BFDSessionSpec struct {
	// PeerAddress is the IP address of the BFD peer.
	// +kubebuilder:validation:Required
	PeerAddress string `json:"peerAddress"`

	// DetectMultiplier is the detection time multiplier. Default: 3.
	// +kubebuilder:default=3
	// +kubebuilder:validation:Minimum=1
	DetectMultiplier int32 `json:"detectMultiplier"`

	// ReceiveInterval is the minimum receive interval in milliseconds. Default: 300.
	// +kubebuilder:default=300
	// +kubebuilder:validation:Minimum=10
	ReceiveInterval int32 `json:"receiveInterval"`

	// TransmitInterval is the minimum transmit interval in milliseconds. Default: 300.
	// +kubebuilder:default=300
	// +kubebuilder:validation:Minimum=10
	TransmitInterval int32 `json:"transmitInterval"`
}

// BFDSessionStatus defines the observed state of a BFD session.
type BFDSessionStatus struct {
	// State is the BFD session state (Up/Down/Init/AdminDown).
	// +optional
	State string `json:"state,omitempty"`

	// ObservedGeneration is the last generation reconciled.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds reconciliation status conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// BFDSession is the Schema for Bidirectional Forwarding Detection sessions.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=bfd,categories=straitgateway
// +kubebuilder:printcolumn:name="PeerAddress",type=string,JSONPath=`.spec.peerAddress`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type BFDSession struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BFDSessionSpec   `json:"spec,omitempty"`
	Status BFDSessionStatus `json:"status,omitempty"`
}

// BFDSessionList contains a list of BFDSession.
// +kubebuilder:object:root=true
type BFDSessionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BFDSession `json:"items"`
}
