package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// GlobalServicePort describes a port exposed in the multi-cluster mesh
type GlobalServicePort struct {
	Name       string `json:"name,omitempty"`
	Protocol   string `json:"protocol,omitempty"` // TCP, UDP
	Port       int32  `json:"port"`
	TargetPort int32  `json:"targetPort,omitempty"`
}

// ClusterBackendRef represents a backend provider cluster for a global service
type ClusterBackendRef struct {
	ClusterID   string   `json:"clusterId"`
	Weight      int32    `json:"weight,omitempty"`
	EndpointIPs []string `json:"endpointIps,omitempty"`
}

// ClusterPeeringSpec defines the desired peering with a remote Kubernetes cluster
type ClusterPeeringSpec struct {
	// ClusterID is the unique identifier for the remote peer cluster
	ClusterID string `json:"clusterId"`

	// APIServerURL is the public/reachable endpoint of the remote Kubernetes API server
	APIServerURL string `json:"apiServerUrl,omitempty"`

	// WireGuardEndpoint is the reachable UDP endpoint (host:port) for encrypted transit
	WireGuardEndpoint string `json:"wireguardEndpoint"`

	// WireGuardPublicKey is the remote cluster's WireGuard public key
	WireGuardPublicKey string `json:"wireguardPublicKey"`

	// PodCIDRs are the pod IP ranges allocated within the remote cluster
	PodCIDRs []string `json:"podCIDRs"`

	// ServiceCIDRs are the service IP ranges allocated within the remote cluster
	ServiceCIDRs []string `json:"serviceCIDRs,omitempty"`

	// VRFID specifies the VRF routing domain for this multi-cluster mesh
	VRFID uint32 `json:"vrfId,omitempty"`

	// PersistentKeepaliveSeconds keeps NAT mappings alive
	PersistentKeepaliveSeconds int32 `json:"persistentKeepaliveSeconds,omitempty"`
}

// ClusterPeeringStatus defines the observed state of ClusterPeering
type ClusterPeeringStatus struct {
	Phase                string             `json:"phase,omitempty"` // Connected, Degraded, Disconnected
	ActiveRoutes         int32              `json:"activeRoutes,omitempty"`
	RemoteEndpointsCount int32              `json:"remoteEndpointsCount,omitempty"`
	LatencyMillis        float64            `json:"latencyMillis,omitempty"`
	LastHeartbeat        *metav1.Time       `json:"lastHeartbeat,omitempty"`
	Conditions           []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ClusterID",type="string",JSONPath=".spec.clusterId"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".spec.wireguardEndpoint"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"

// ClusterPeering is the Schema for multi-cluster peering connections
type ClusterPeering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterPeeringSpec   `json:"spec,omitempty"`
	Status ClusterPeeringStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterPeeringList contains a list of ClusterPeering
type ClusterPeeringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterPeering `json:"items"`
}

// GlobalServiceExportSpec defines which local service is exported across ClusterMesh
type GlobalServiceExportSpec struct {
	ServiceName      string              `json:"serviceName"`
	ServiceNamespace string              `json:"serviceNamespace"`
	ExportedPorts    []GlobalServicePort `json:"exportedPorts,omitempty"`
	TrafficPolicy    string              `json:"trafficPolicy,omitempty"` // ClusterMeshPreferred, ClusterLocal, RoundRobin
}

// GlobalServiceExportStatus defines export state
type GlobalServiceExportStatus struct {
	Exported        bool         `json:"exported,omitempty"`
	ActiveEndpoints int32        `json:"activeEndpoints,omitempty"`
	ExportedAt      *metav1.Time `json:"exportedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GlobalServiceExport is the Schema for exporting local services across the multi-cluster mesh
type GlobalServiceExport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlobalServiceExportSpec   `json:"spec,omitempty"`
	Status GlobalServiceExportStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GlobalServiceExportList contains a list of GlobalServiceExport
type GlobalServiceExportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalServiceExport `json:"items"`
}

// GlobalServiceImportSpec defines the imported multi-cluster service
type GlobalServiceImportSpec struct {
	GlobalServiceName string              `json:"globalServiceName"` // e.g. catalog.global.straitgateway
	GlobalVIP         string              `json:"globalVip,omitempty"`
	Ports             []GlobalServicePort `json:"ports,omitempty"`
	Clusters          []ClusterBackendRef `json:"clusters,omitempty"`
}

// GlobalServiceImportStatus defines import state
type GlobalServiceImportStatus struct {
	EndpointsCount int32              `json:"endpointsCount,omitempty"`
	Ready          bool               `json:"ready,omitempty"`
	Conditions     []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GlobalServiceImport is the Schema for importing multi-cluster services
type GlobalServiceImport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlobalServiceImportSpec   `json:"spec,omitempty"`
	Status GlobalServiceImportStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GlobalServiceImportList contains a list of GlobalServiceImport
type GlobalServiceImportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalServiceImport `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&ClusterPeering{}, &ClusterPeeringList{},
		&GlobalServiceExport{}, &GlobalServiceExportList{},
		&GlobalServiceImport{}, &GlobalServiceImportList{},
	)
}

// DeepCopyInto for ClusterPeering
func (in *ClusterPeering) DeepCopyInto(out *ClusterPeering) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *ClusterPeering) DeepCopy() *ClusterPeering {
	if in == nil {
		return nil
	}
	out := new(ClusterPeering)
	in.DeepCopyInto(out)
	return out
}

func (in *ClusterPeering) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *ClusterPeeringList) DeepCopyInto(out *ClusterPeeringList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterPeering, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *ClusterPeeringList) DeepCopy() *ClusterPeeringList {
	if in == nil {
		return nil
	}
	out := new(ClusterPeeringList)
	in.DeepCopyInto(out)
	return out
}

func (in *ClusterPeeringList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *ClusterPeeringSpec) DeepCopyInto(out *ClusterPeeringSpec) {
	*out = *in
	if in.PodCIDRs != nil {
		in, out := &in.PodCIDRs, &out.PodCIDRs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ServiceCIDRs != nil {
		in, out := &in.ServiceCIDRs, &out.ServiceCIDRs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

func (in *ClusterPeeringStatus) DeepCopyInto(out *ClusterPeeringStatus) {
	*out = *in
	if in.LastHeartbeat != nil {
		in, out := &in.LastHeartbeat, &out.LastHeartbeat
		*out = (*in).DeepCopy()
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy for GlobalServiceExport
func (in *GlobalServiceExport) DeepCopyInto(out *GlobalServiceExport) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *GlobalServiceExport) DeepCopy() *GlobalServiceExport {
	if in == nil {
		return nil
	}
	out := new(GlobalServiceExport)
	in.DeepCopyInto(out)
	return out
}

func (in *GlobalServiceExport) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *GlobalServiceExportList) DeepCopyInto(out *GlobalServiceExportList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GlobalServiceExport, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *GlobalServiceExportList) DeepCopy() *GlobalServiceExportList {
	if in == nil {
		return nil
	}
	out := new(GlobalServiceExportList)
	in.DeepCopyInto(out)
	return out
}

func (in *GlobalServiceExportList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *GlobalServiceExportSpec) DeepCopyInto(out *GlobalServiceExportSpec) {
	*out = *in
	if in.ExportedPorts != nil {
		in, out := &in.ExportedPorts, &out.ExportedPorts
		*out = make([]GlobalServicePort, len(*in))
		copy(*out, *in)
	}
}

func (in *GlobalServiceExportStatus) DeepCopyInto(out *GlobalServiceExportStatus) {
	*out = *in
	if in.ExportedAt != nil {
		in, out := &in.ExportedAt, &out.ExportedAt
		*out = (*in).DeepCopy()
	}
}

// DeepCopy for GlobalServiceImport
func (in *GlobalServiceImport) DeepCopyInto(out *GlobalServiceImport) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *GlobalServiceImport) DeepCopy() *GlobalServiceImport {
	if in == nil {
		return nil
	}
	out := new(GlobalServiceImport)
	in.DeepCopyInto(out)
	return out
}

func (in *GlobalServiceImport) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *GlobalServiceImportList) DeepCopyInto(out *GlobalServiceImportList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GlobalServiceImport, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *GlobalServiceImportList) DeepCopy() *GlobalServiceImportList {
	if in == nil {
		return nil
	}
	out := new(GlobalServiceImportList)
	in.DeepCopyInto(out)
	return out
}

func (in *GlobalServiceImportList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *GlobalServiceImportSpec) DeepCopyInto(out *GlobalServiceImportSpec) {
	*out = *in
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]GlobalServicePort, len(*in))
		copy(*out, *in)
	}
	if in.Clusters != nil {
		in, out := &in.Clusters, &out.Clusters
		*out = make([]ClusterBackendRef, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *ClusterBackendRef) DeepCopyInto(out *ClusterBackendRef) {
	*out = *in
	if in.EndpointIPs != nil {
		in, out := &in.EndpointIPs, &out.EndpointIPs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

func (in *GlobalServiceImportStatus) DeepCopyInto(out *GlobalServiceImportStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}
