package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DataplaneMode defines the dataplane hook mode
type DataplaneMode string

const (
	ModeTCX DataplaneMode = "TCX"
	ModeTC  DataplaneMode = "TC"
	ModeXDP DataplaneMode = "XDP"
)

// StraitGatewayConfigSpec defines cluster-wide StraitGateway settings
type StraitGatewayConfigSpec struct {
	// ClusterCIDR is the global pod IP range (e.g. "10.244.0.0/16")
	ClusterCIDR string `json:"clusterCidr"`

	// ServiceCIDR is the Kubernetes service VIP range (e.g. "10.96.0.0/12")
	ServiceCIDR string `json:"serviceCidr"`

	// DataplaneMode specifies whether to use TCX, standard TC, or XDP fastpath
	// +kubebuilder:default="TCX"
	DataplaneMode DataplaneMode `json:"dataplaneMode"`

	// TunnelType specifies inter-node encapsulation ("WireGuard" or "DirectRouting")
	// +kubebuilder:default="WireGuard"
	TunnelType string `json:"tunnelType"`

	// EnableSocketLB enables socket-level cgroup load balancing for kube-proxy replacement
	// +kubebuilder:default=true
	EnableSocketLB bool `json:"enableSocketLb"`

	// EnableMaglev enables Maglev consistent hashing for service load balancing
	// +kubebuilder:default=true
	EnableMaglev bool `json:"enableMaglev"`

	// EnableObservability activates BPF ring buffer telemetry export
	// +kubebuilder:default=true
	EnableObservability bool `json:"enableObservability"`
}

// StraitGatewayConfigStatus defines observed cluster network state
type StraitGatewayConfigStatus struct {
	// Ready indicates whether all node agents report healthy dataplane state
	Ready bool `json:"ready"`

	// ActiveNodes is the number of node agents connected and reconciled
	ActiveNodes int32 `json:"activeNodes"`

	// Conditions represent high-level cluster networking status
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Mode",type="string",JSONPath=".spec.dataplaneMode"
// +kubebuilder:printcolumn:name="SocketLB",type="boolean",JSONPath=".spec.enableSocketLb"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Nodes",type="integer",JSONPath=".status.activeNodes"

// StraitGatewayConfig is the Schema for the straitgatewayconfigs API
type StraitGatewayConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StraitGatewayConfigSpec   `json:"spec,omitempty"`
	Status StraitGatewayConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StraitGatewayConfigList contains a list of StraitGatewayConfig
type StraitGatewayConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StraitGatewayConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StraitGatewayConfig{}, &StraitGatewayConfigList{})
}
