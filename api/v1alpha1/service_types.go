package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MTLSMode defines service encryption state
type MTLSMode string

const (
	MTLSStrict     MTLSMode = "Strict"
	MTLSPermissive MTLSMode = "Permissive"
	MTLSDisabled   MTLSMode = "Disabled"
)

// ServiceMeshPolicySpec defines the desired state of ServiceMeshPolicy
type ServiceMeshPolicySpec struct {
	// ServiceRef references the target Kubernetes Service
	ServiceRef string `json:"serviceRef"`

	// MTLS specifies mutual TLS configuration
	// +optional
	MTLS *MTLSConfig `json:"mtls,omitempty"`

	// TrafficSplit specifies weighted routing across multiple backend versions
	// +optional
	TrafficSplit []TrafficSplitTarget `json:"trafficSplit,omitempty"`

	// Retries specifies automatic HTTP retry behavior
	// +optional
	Retries *RetryPolicy `json:"retries,omitempty"`

	// TimeoutSeconds specifies request timeout in seconds
	// +optional
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`

	// CircuitBreaker specifies failure thresholds
	// +optional
	CircuitBreaker *CircuitBreakerConfig `json:"circuitBreaker,omitempty"`
}

// MTLSConfig configures workload encryption
type MTLSConfig struct {
	// Mode is Strict, Permissive, or Disabled
	Mode MTLSMode `json:"mode"`

	// TrustDomain specifies SPIFFE identity domain (e.g. "cluster.local")
	// +optional
	TrustDomain string `json:"trustDomain,omitempty"`
}

// TrafficSplitTarget holds weighted backend destinations
type TrafficSplitTarget struct {
	// BackendService references the destination service or subset
	BackendService string `json:"backendService"`

	// Weight is relative proportion of traffic (0-100)
	Weight int32 `json:"weight"`
}

// RetryPolicy specifies retry conditions
type RetryPolicy struct {
	// Attempts is max number of retries
	Attempts int32 `json:"attempts"`

	// PerTryTimeoutSeconds specifies timeout per attempt
	// +optional
	PerTryTimeoutSeconds int32 `json:"perTryTimeoutSeconds,omitempty"`

	// RetryOn specifies HTTP status codes or events (e.g. "5xx", "connect-failure")
	// +optional
	RetryOn string `json:"retryOn,omitempty"`
}

// CircuitBreakerConfig specifies fault isolation
type CircuitBreakerConfig struct {
	// MaxConnections is the maximum number of concurrent connections
	// +optional
	MaxConnections int32 `json:"maxConnections,omitempty"`

	// MaxPendingRequests is max requests queued waiting for a connection
	// +optional
	MaxPendingRequests int32 `json:"maxPendingRequests,omitempty"`

	// ConsecutiveErrors before tripping the circuit breaker
	// +optional
	ConsecutiveErrors int32 `json:"consecutiveErrors,omitempty"`
}

// ServiceMeshPolicyStatus defines the observed state of ServiceMeshPolicy
type ServiceMeshPolicyStatus struct {
	// Phase is "Active", "Pending", or "Error"
	Phase string `json:"phase,omitempty"`

	// EndpointsActive is the count of active endpoints configured in BPF maps
	EndpointsActive int32 `json:"endpointsActive"`

	// Conditions represent status observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Service",type="string",JSONPath=".spec.serviceRef"
// +kubebuilder:printcolumn:name="mTLS",type="string",JSONPath=".spec.mtls.mode"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"

// ServiceMeshPolicy is the Schema for the servicemeshpolicies API
type ServiceMeshPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceMeshPolicySpec   `json:"spec,omitempty"`
	Status ServiceMeshPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceMeshPolicyList contains a list of ServiceMeshPolicy
type ServiceMeshPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceMeshPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceMeshPolicy{}, &ServiceMeshPolicyList{})
}
