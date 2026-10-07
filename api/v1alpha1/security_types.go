package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PolicyAction defines enforcement response
type PolicyAction string

const (
	ActionAllow PolicyAction = "Allow"
	ActionDeny  PolicyAction = "Deny"
	ActionAudit PolicyAction = "Audit"
)

// SecurityPolicySpec defines the desired state of SecurityPolicy
type SecurityPolicySpec struct {
	// Selector selects the pods to which this policy applies
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Ingress rules evaluated for inbound traffic to selected pods
	// +optional
	Ingress []RuleIngress `json:"ingress,omitempty"`

	// Egress rules evaluated for outbound traffic from selected pods
	// +optional
	Egress []RuleEgress `json:"egress,omitempty"`

	// DefaultAction specifies default action if no rule matches ("Allow" or "Deny")
	// +kubebuilder:default="Deny"
	DefaultAction PolicyAction `json:"defaultAction"`

	// Priority defines the evaluation precedence (lower numbers evaluate first)
	// +optional
	Priority int32 `json:"priority,omitempty"`
}

// RuleIngress describes an ingress permission rule
type RuleIngress struct {
	// From specifies sources allowed to reach selected workloads
	// +optional
	From []RulePeer `json:"from,omitempty"`

	// Ports lists destination ports and protocols permitted
	// +optional
	Ports []SecurityPort `json:"ports,omitempty"`

	// HTTP specifies Layer-7 HTTP rules (for service mesh enforcement)
	// +optional
	HTTP []HTTPRule `json:"http,omitempty"`
}

// RuleEgress describes an egress permission rule
type RuleEgress struct {
	// To specifies destinations permitted from selected workloads
	// +optional
	To []RulePeer `json:"to,omitempty"`

	// Ports lists destination ports and protocols permitted
	// +optional
	Ports []SecurityPort `json:"ports,omitempty"`
}

// RulePeer specifies source or destination entity selectors
type RulePeer struct {
	// PodSelector selects pods in the same namespace (or selected namespaces)
	// +optional
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty"`

	// NamespaceSelector selects namespaces
	// +optional
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`

	// CIDR specifies an IP prefix for non-cluster traffic
	// +optional
	CIDR string `json:"cidr,omitempty"`

	// ServiceAccount matches workloads by service account name
	// +optional
	ServiceAccount string `json:"serviceAccount,omitempty"`
}

// SecurityPort specifies transport layer port and protocol
type SecurityPort struct {
	// Protocol is TCP, UDP, or SCTP
	Protocol string `json:"protocol"`

	// Port number
	Port int32 `json:"port"`

	// EndPort defines a port range [Port, EndPort]
	// +optional
	EndPort int32 `json:"endPort,omitempty"`
}

// HTTPRule specifies L7 path and method constraints
type HTTPRule struct {
	// Methods lists allowed HTTP methods (e.g. GET, POST)
	// +optional
	Methods []string `json:"methods,omitempty"`

	// Paths lists exact or prefix path matches
	// +optional
	Paths []string `json:"paths,omitempty"`
}

// SecurityPolicyStatus defines the observed state of SecurityPolicy
type SecurityPolicyStatus struct {
	// ActiveNodes lists nodes where this policy is compiled into BPF maps
	// +optional
	ActiveNodes []string `json:"activeNodes,omitempty"`

	// MatchedPods counts current workloads bound to this policy
	MatchedPods int32 `json:"matchedPods"`

	// Phase reports the compilation status ("Ready", "Error")
	Phase string `json:"phase,omitempty"`

	// Conditions represent status observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Action",type="string",JSONPath=".spec.defaultAction"
// +kubebuilder:printcolumn:name="Priority",type="integer",JSONPath=".spec.priority"
// +kubebuilder:printcolumn:name="Pods",type="integer",JSONPath=".status.matchedPods"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"

// SecurityPolicy is the Schema for the securitypolicies API
type SecurityPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecurityPolicySpec   `json:"spec,omitempty"`
	Status SecurityPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecurityPolicyList contains a list of SecurityPolicy
type SecurityPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecurityPolicy{}, &SecurityPolicyList{})
}
