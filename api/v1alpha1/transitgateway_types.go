package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TransitGatewaySpec defines the desired state of TransitGateway
type TransitGatewaySpec struct {
	// AutonomousSystemNumber (ASN) is the local BGP ASN for this Transit Gateway
	ASN uint32 `json:"asn"`

	// RouterID is the 32-bit IPv4 identifier for BGP routing
	RouterID string `json:"routerId"`

	// MTU is the Maximum Transmission Unit for transit dataplane interfaces
	// +optional
	MTU int32 `json:"mtu,omitempty"`

	// VRFs defines the isolated routing domains
	// +optional
	VRFs []VRFConfig `json:"vrfs,omitempty"`

	// WireGuard configures the cluster overlay tunnel endpoint
	// +optional
	WireGuard *WireGuardConfig `json:"wireguard,omitempty"`

	// BFD configures the Bidirectional Forwarding Detection default profile
	// +optional
	BFD *BFDConfig `json:"bfd,omitempty"`

	// DefaultAction specifies default transit policy ("Allow" or "Deny")
	// +optional
	DefaultAction string `json:"defaultAction,omitempty"`
}

// VRFConfig defines an individual Virtual Routing and Forwarding domain
type VRFConfig struct {
	// ID is the numeric identifier for the VRF (0 is default)
	ID uint32 `json:"id"`

	// Name is the human-readable label for this VRF
	Name string `json:"name"`

	// Description provides contextual details
	// +optional
	Description string `json:"description,omitempty"`

	// ExportRouteTargets lists BGP route target communities exported from this VRF
	// +optional
	ExportRouteTargets []string `json:"exportRouteTargets,omitempty"`

	// ImportRouteTargets lists BGP route target communities imported into this VRF
	// +optional
	ImportRouteTargets []string `json:"importRouteTargets,omitempty"`
}

// WireGuardConfig contains tunnel endpoint settings
type WireGuardConfig struct {
	// ListenPort is the UDP port for WireGuard tunnels (default: 51820)
	// +optional
	ListenPort int32 `json:"listenPort,omitempty"`

	// PublicKey is the active WireGuard public key for this gateway
	// +optional
	PublicKey string `json:"publicKey,omitempty"`

	// PrivateKeySecretRef references the Secret containing the private key
	// +optional
	PrivateKeySecretRef string `json:"privateKeySecretRef,omitempty"`
}

// BFDConfig contains BFD session parameters
type BFDConfig struct {
	// MinTxIntervalMs is the minimum interval in milliseconds between transmitting BFD control packets
	// +optional
	MinTxIntervalMs uint32 `json:"minTxIntervalMs,omitempty"`

	// MinRxIntervalMs is the minimum interval in milliseconds between receiving BFD control packets
	// +optional
	MinRxIntervalMs uint32 `json:"minRxIntervalMs,omitempty"`

	// DetectMultiplier is the number of missed packets before declaring session down
	// +optional
	DetectMultiplier uint8 `json:"detectMultiplier,omitempty"`
}

// TransitGatewayStatus defines the observed state of TransitGateway
type TransitGatewayStatus struct {
	// Phase indicates the overall operational lifecycle phase (Pending, Ready, Error)
	Phase string `json:"phase,omitempty"`

	// ActivePeers counts the currently established BGP/WireGuard peers
	ActivePeers int32 `json:"activePeers,omitempty"`

	// ActiveRoutes counts the routes actively installed into the BPF route map
	ActiveRoutes int32 `json:"activeRoutes,omitempty"`

	// Conditions represent the latest available observations of object's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ASN",type="integer",JSONPath=".spec.asn"
// +kubebuilder:printcolumn:name="RouterID",type="string",JSONPath=".spec.routerId"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Peers",type="integer",JSONPath=".status.activePeers"
// +kubebuilder:printcolumn:name="Routes",type="integer",JSONPath=".status.activeRoutes"

// TransitGateway is the Schema for the transitgateways API
type TransitGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransitGatewaySpec   `json:"spec,omitempty"`
	Status TransitGatewayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TransitGatewayList contains a list of TransitGateway
type TransitGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitGateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TransitGateway{}, &TransitGatewayList{})
}
