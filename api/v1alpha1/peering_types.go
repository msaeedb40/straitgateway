package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GatewayPeerSpec defines the desired state of GatewayPeer
type GatewayPeerSpec struct {
	// TransitGatewayRef references the parent TransitGateway resource
	TransitGatewayRef string `json:"transitGatewayRef"`

	// PeerType specifies the type of peer ("BGP", "WireGuard", or "Dual")
	PeerType string `json:"peerType"`

	// PeerIP is the IP address of the remote neighbor
	PeerIP string `json:"peerIp"`

	// PeerASN is the Autonomous System Number of the remote BGP peer
	// +optional
	PeerASN uint32 `json:"peerAsn,omitempty"`

	// HoldTimeSeconds is the BGP hold time (default: 90s)
	// +optional
	HoldTimeSeconds uint32 `json:"holdTimeSeconds,omitempty"`

	// KeepaliveSeconds is the BGP keepalive interval (default: 30s)
	// +optional
	KeepaliveSeconds uint32 `json:"keepaliveSeconds,omitempty"`

	// AuthPasswordSecretRef points to Secret with MD5 password for BGP session
	// +optional
	AuthPasswordSecretRef string `json:"authPasswordSecretRef,omitempty"`

	// EnableBFD activates sub-second failure detection for this peer
	// +optional
	EnableBFD bool `json:"enableBfd,omitempty"`

	// WireGuardPeer specifies the cryptographic parameters if using encrypted transit
	// +optional
	WireGuardPeer *WireGuardPeerConfig `json:"wireguardPeer,omitempty"`

	// VRF specifies the routing domain binding for this peer (default: 0)
	// +optional
	VRF uint32 `json:"vrf,omitempty"`

	// AdvertisedPrefixes lists CIDRs to advertise specifically to this peer
	// +optional
	AdvertisedPrefixes []string `json:"advertisedPrefixes,omitempty"`
}

// WireGuardPeerConfig holds the remote WireGuard endpoint parameters
type WireGuardPeerConfig struct {
	// PublicKey is the remote peer's WireGuard public key
	PublicKey string `json:"publicKey"`

	// Endpoint is the remote host:port for UDP tunnel transport
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// AllowedIPs is the list of CIDRs routed through this WireGuard peer
	AllowedIPs []string `json:"allowedIPs"`

	// PersistentKeepaliveSeconds keeps NAT mappings alive
	// +optional
	PersistentKeepaliveSeconds int32 `json:"persistentKeepaliveSeconds,omitempty"`
}

// GatewayPeerStatus defines the observed state of GatewayPeer
type GatewayPeerStatus struct {
	// SessionState indicates the protocol state ("Established", "Active", "Down", "Idle")
	SessionState string `json:"sessionState,omitempty"`

	// BFDState indicates the BFD status ("Up", "Down", "AdminDown", "Disabled")
	// +optional
	BFDState string `json:"bfdState,omitempty"`

	// ReceivedRoutes counts routes received from this peer
	ReceivedRoutes int32 `json:"receivedRoutes,omitempty"`

	// AdvertisedRoutes counts routes advertised to this peer
	AdvertisedRoutes int32 `json:"advertisedRoutes,omitempty"`

	// LastStateChange records the timestamp of the last status transition
	// +optional
	LastStateChange *metav1.Time `json:"lastStateChange,omitempty"`

	// Conditions represent observations of peer status
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.peerType"
// +kubebuilder:printcolumn:name="PeerIP",type="string",JSONPath=".spec.peerIp"
// +kubebuilder:printcolumn:name="ASN",type="integer",JSONPath=".spec.peerAsn"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.sessionState"
// +kubebuilder:printcolumn:name="BFD",type="string",JSONPath=".status.bfdState"

// GatewayPeer is the Schema for the gatewaypeers API
type GatewayPeer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatewayPeerSpec   `json:"spec,omitempty"`
	Status GatewayPeerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GatewayPeerList contains a list of GatewayPeer
type GatewayPeerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GatewayPeer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GatewayPeer{}, &GatewayPeerList{})
}
