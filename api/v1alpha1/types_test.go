package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestSchemeRegistration(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add v1alpha1 to scheme: %v", err)
	}

	types := []runtime.Object{
		&TransitGateway{},
		&TransitGatewayList{},
		&GatewayPeer{},
		&GatewayPeerList{},
		&TransitRoute{},
		&TransitRouteList{},
		&SecurityPolicy{},
		&SecurityPolicyList{},
		&ServiceMeshPolicy{},
		&ServiceMeshPolicyList{},
		&StraitGatewayConfig{},
		&StraitGatewayConfigList{},
		&ClusterPeering{},
		&ClusterPeeringList{},
		&GlobalServiceExport{},
		&GlobalServiceExportList{},
		&GlobalServiceImport{},
		&GlobalServiceImportList{},
	}

	for _, obj := range types {
		gvks, unversioned, err := scheme.ObjectKinds(obj)
		if err != nil {
			t.Errorf("expected object %T to be registered in scheme, got error: %v", obj, err)
		}
		if unversioned {
			t.Errorf("object %T should not be unversioned", obj)
		}
		if len(gvks) == 0 {
			t.Errorf("object %T had no GVKs registered", obj)
		}
	}
}

func TestDeepCopyTransitGateway(t *testing.T) {
	tg := &TransitGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tgw-primary",
		},
		Spec: TransitGatewaySpec{
			ASN:      65001,
			RouterID: "10.0.0.1",
			VRFs: []VRFConfig{
				{
					ID:                 100,
					Name:               "tenant-a",
					ExportRouteTargets: []string{"65001:100"},
				},
			},
		},
	}

	copied := tg.DeepCopy()
	if copied == nil {
		t.Fatal("expected non-nil deepcopy")
	}
	if copied.Spec.ASN != tg.Spec.ASN {
		t.Errorf("expected ASN %d, got %d", tg.Spec.ASN, copied.Spec.ASN)
	}
	if len(copied.Spec.VRFs) != 1 || copied.Spec.VRFs[0].Name != "tenant-a" {
		t.Errorf("VRFs slice was not properly deepcopied: %+v", copied.Spec.VRFs)
	}

	// Ensure modifying copy doesn't mutate original
	copied.Spec.VRFs[0].Name = "mutated"
	if tg.Spec.VRFs[0].Name == "mutated" {
		t.Error("mutation in copy affected the original object")
	}
}

func TestDeepCopySecurityPolicy(t *testing.T) {
	sp := &SecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-policy",
			Namespace: "default",
		},
		Spec: SecurityPolicySpec{
			DefaultAction: ActionDeny,
			Priority:      10,
			Ingress: []RuleIngress{
				{
					Ports: []SecurityPort{
						{Protocol: "TCP", Port: 8080},
					},
					HTTP: []HTTPRule{
						{Methods: []string{"GET"}, Paths: []string{"/api"}},
					},
				},
			},
		},
	}

	copied := sp.DeepCopy()
	if copied == nil {
		t.Fatal("expected non-nil deepcopy")
	}
	if copied.Spec.DefaultAction != ActionDeny {
		t.Errorf("expected DefaultAction %s, got %s", ActionDeny, copied.Spec.DefaultAction)
	}
	if len(copied.Spec.Ingress[0].HTTP[0].Methods) != 1 {
		t.Errorf("HTTP methods slice not copied properly")
	}
}

func TestDeepCopyClusterMesh(t *testing.T) {
	cp := &ClusterPeering{
		ObjectMeta: metav1.ObjectMeta{
			Name: "peer-west",
		},
		Spec: ClusterPeeringSpec{
			ClusterID:         "cluster-west-1",
			WireGuardEndpoint: "198.51.100.1:51820",
			PodCIDRs:          []string{"10.244.0.0/16"},
			ServiceCIDRs:      []string{"10.96.0.0/16"},
		},
	}
	cpCopy := cp.DeepCopy()
	if cpCopy == nil || cpCopy.Spec.ClusterID != "cluster-west-1" {
		t.Fatalf("failed deepcopy on ClusterPeering")
	}
	cpCopy.Spec.PodCIDRs[0] = "10.0.0.0/16"
	if cp.Spec.PodCIDRs[0] == "10.0.0.0/16" {
		t.Errorf("expected deepcopy slice isolation")
	}

	gsi := &GlobalServiceImport{
		ObjectMeta: metav1.ObjectMeta{
			Name: "order-service",
		},
		Spec: GlobalServiceImportSpec{
			GlobalServiceName: "order.global.straitgateway",
			GlobalVIP:         "240.0.0.50",
			Clusters: []ClusterBackendRef{
				{ClusterID: "cluster-west-1", EndpointIPs: []string{"10.244.1.20"}},
			},
		},
	}
	gsiCopy := gsi.DeepCopy()
	if gsiCopy == nil || gsiCopy.Spec.GlobalVIP != "240.0.0.50" {
		t.Fatalf("failed deepcopy on GlobalServiceImport")
	}
}
