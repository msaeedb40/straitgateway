// Package v1alpha4 defines the StraitGateway API types for Kubernetes CRDs.
package v1alpha4

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GroupName    = "straitgateway.io"
	GroupVersion = "v1alpha4"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// StraitGatewayConfig is the top-level configuration resource.
type StraitGatewayConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              StraitGatewayConfigSpec   `json:"spec,omitempty"`
	Status            StraitGatewayConfigStatus `json:"status,omitempty"`
}

type StraitGatewayConfigSpec struct {
	CNI           CNIConfig           `json:"cni,omitempty"`
	IPAM          IPAMConfig          `json:"ipam,omitempty"`
	Datapath      DatapathConfig      `json:"datapath,omitempty"`
	Service       ServiceConfig       `json:"service,omitempty"`
	Policy        PolicyConfig        `json:"policy,omitempty"`
	Transit       TransitConfig       `json:"transit,omitempty"`
	Mesh          MeshConfig          `json:"mesh,omitempty"`
	Observability ObservabilityConfig `json:"observability,omitempty"`
}

type CNIConfig struct {
	Enabled bool `json:"enabled"`
	MTU     int  `json:"mtu,omitempty"`
}

type IPAMConfig struct {
	Mode         string   `json:"mode,omitempty"`
	PodCIDRs     []string `json:"podCIDRs,omitempty"`
	ServiceCIDRs []string `json:"serviceCIDRs,omitempty"`
}

type DatapathConfig struct {
	Mode      string `json:"mode,omitempty"` // "netkit" (default) or "veth"
	EnableXDP bool   `json:"enableXDP,omitempty"`
	EnableTCX bool   `json:"enableTCX,omitempty"`
}

type ServiceConfig struct {
	Enabled               bool `json:"enabled"`
	EnableNodePort        bool `json:"enableNodePort,omitempty"`
	EnableExternalIPs     bool `json:"enableExternalIPs,omitempty"`
	EnableSessionAffinity bool `json:"enableSessionAffinity,omitempty"`
}

type PolicyConfig struct{ Enabled bool `json:"enabled"` }

type TransitConfig struct {
	Enabled  bool   `json:"enabled"`
	Topology string `json:"topology,omitempty"` // hub-spoke|mesh|peer-to-peer|hub-to-hub|hybrid
}

type MeshConfig struct{ Enabled bool `json:"enabled"` }

type ObservabilityConfig struct {
	EnablePrometheus    bool `json:"enablePrometheus,omitempty"`
	EnableOpenTelemetry bool `json:"enableOpenTelemetry,omitempty"`
}

type StraitGatewayConfigStatus struct {
	Phase          string             `json:"phase,omitempty"`
	Conditions     []metav1.Condition `json:"conditions,omitempty"`
	NodeCount      int32              `json:"nodeCount,omitempty"`
	ReadyNodeCount int32              `json:"readyNodeCount,omitempty"`
}

// +kubebuilder:object:root=true
type StraitGatewayConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StraitGatewayConfig `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

// StraitGatewayNode represents per-node StraitGateway state.
type StraitGatewayNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              StraitGatewayNodeSpec   `json:"spec,omitempty"`
	Status            StraitGatewayNodeStatus `json:"status,omitempty"`
}

type StraitGatewayNodeSpec struct {
	NodeName string   `json:"nodeName"`
	PodCIDRs []string `json:"podCIDRs,omitempty"`
}

type StraitGatewayNodeStatus struct {
	Phase            string             `json:"phase,omitempty"`
	DatapathReady    bool               `json:"datapathReady,omitempty"`
	CNIReady         bool               `json:"cniReady,omitempty"`
	EBPFProgramCount int32              `json:"ebpfProgramCount,omitempty"`
	EBPFMapCount     int32              `json:"ebpfMapCount,omitempty"`
	Conditions       []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
type StraitGatewayNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StraitGatewayNode `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced

// StraitGatewayPolicy defines a network policy resource.
type StraitGatewayPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              StraitGatewayPolicySpec   `json:"spec,omitempty"`
	Status            StraitGatewayPolicyStatus `json:"status,omitempty"`
}

type StraitGatewayPolicySpec struct {
	Selector metav1.LabelSelector `json:"selector"`
	Ingress  []PolicyRule         `json:"ingress,omitempty"`
	Egress   []PolicyRule         `json:"egress,omitempty"`
}

type PolicyRule struct {
	Ports  []PolicyPort `json:"ports,omitempty"`
	From   []PolicyPeer `json:"from,omitempty"`
	To     []PolicyPeer `json:"to,omitempty"`
	Action string       `json:"action"` // Allow | Deny
}

type PolicyPort struct {
	Protocol string `json:"protocol,omitempty"`
	Port     int32  `json:"port,omitempty"`
	EndPort  int32  `json:"endPort,omitempty"`
}

type PolicyPeer struct {
	PodSelector       *metav1.LabelSelector `json:"podSelector,omitempty"`
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`
	CIDRs             []string              `json:"cidrs,omitempty"`
}

type StraitGatewayPolicyStatus struct {
	Phase      string             `json:"phase,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
type StraitGatewayPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StraitGatewayPolicy `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// TransitGateway defines a transit gateway resource.
type TransitGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TransitGatewaySpec   `json:"spec,omitempty"`
	Status            TransitGatewayStatus `json:"status,omitempty"`
}

type TransitGatewaySpec struct {
	Topology string           `json:"topology"`
	Networks []TransitNetwork `json:"networks,omitempty"`
}

type TransitNetwork struct {
	Name    string   `json:"name"`
	CIDRs   []string `json:"cidrs,omitempty"`
	Segment string   `json:"segment,omitempty"`
}

type TransitGatewayStatus struct {
	Phase        string             `json:"phase,omitempty"`
	PeerCount    int32              `json:"peerCount,omitempty"`
	NetworkCount int32              `json:"networkCount,omitempty"`
	Conditions   []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
type TransitGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitGateway `json:"items"`
}
