package peering

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/internal/mesh/clustermesh"
	"github.com/straitgateway/straitgateway/internal/routing"
	"github.com/straitgateway/straitgateway/internal/wireguard"
)

func TestClusterPeeringReconciler(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	peer := &v1alpha1.ClusterPeering{
		ObjectMeta: metav1.ObjectMeta{
			Name: "peering-cluster-2",
		},
		Spec: v1alpha1.ClusterPeeringSpec{
			ClusterID:          "cluster-2",
			WireGuardEndpoint:  "198.51.100.50:51820",
			WireGuardPublicKey: "dummy-pubkey-cl2=",
			PodCIDRs:           []string{"10.244.0.0/16"},
			ServiceCIDRs:       []string{"10.96.0.0/16"},
		},
	}

	fakeClient := clientfake.NewClientBuilder().WithScheme(scheme).WithObjects(peer).Build()

	fib := routing.NewFIBEngine()
	rm := routing.NewRouteManager(fib, nil)
	wg, err := wireguard.NewManager("wg-test", 51820)
	if err != nil {
		t.Fatalf("failed to init wg: %v", err)
	}
	wg.SetRouteManager(rm)

	cmMgr, err := clustermesh.NewManager(clustermesh.DefaultConfig(), wg, rm)
	if err != nil {
		t.Fatalf("failed to init clustermesh: %v", err)
	}

	reconciler := &ClusterPeeringReconciler{
		Client:         fakeClient,
		Scheme:         scheme,
		ClusterMeshMgr: cmMgr,
	}

	req := ctrl.Request{
		NamespacedName: k8stypes.NamespacedName{
			Name: "peering-cluster-2",
		},
	}

	res, err := reconciler.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}
	if res.Requeue { //nolint:staticcheck // SA1019: deprecated field, use RequeueAfter in future
		t.Errorf("expected no requeue")
	}

	// Verify cluster is registered in clustermesh manager
	st, ok := cmMgr.GetClusterPeer("cluster-2")
	if !ok || st.Phase != "Connected" {
		t.Fatalf("expected cluster-2 in Connected phase, got %+v", st)
	}
}

func TestGlobalServiceReconciler(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	export := &v1alpha1.GlobalServiceExport{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "export-payment",
			Namespace: "default",
		},
		Spec: v1alpha1.GlobalServiceExportSpec{
			ServiceName:      "payment",
			ServiceNamespace: "default",
			TrafficPolicy:    "ClusterMeshPreferred",
		},
	}

	importSvc := &v1alpha1.GlobalServiceImport{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "import-auth",
			Namespace: "default",
		},
		Spec: v1alpha1.GlobalServiceImportSpec{
			GlobalServiceName: "auth.default.global.straitgateway",
			GlobalVIP:         "240.0.5.10",
			Clusters: []v1alpha1.ClusterBackendRef{
				{ClusterID: "cluster-2", EndpointIPs: []string{"10.244.5.1"}},
			},
		},
	}

	fakeClient := clientfake.NewClientBuilder().WithScheme(scheme).WithObjects(export, importSvc).Build()

	cmMgr, err := clustermesh.NewManager(clustermesh.DefaultConfig(), nil, nil)
	if err != nil {
		t.Fatalf("failed to init clustermesh: %v", err)
	}

	reconciler := &GlobalServiceReconciler{
		Client:         fakeClient,
		Scheme:         scheme,
		ClusterMeshMgr: cmMgr,
	}

	// Reconcile Export
	_, err = reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: k8stypes.NamespacedName{
			Namespace: "default",
			Name:      "export-payment",
		},
	})
	if err != nil {
		t.Fatalf("export reconcile failed: %v", err)
	}

	// Reconcile Import
	_, err = reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: k8stypes.NamespacedName{
			Namespace: "default",
			Name:      "import-auth",
		},
	})
	if err != nil {
		t.Fatalf("import reconcile failed: %v", err)
	}

	// Verify VIP resolution
	vip, ok := cmMgr.ResolveGlobalVIP("auth.default.global.straitgateway")
	if !ok || vip.String() != "240.0.5.10" {
		t.Errorf("expected 240.0.5.10 VIP, got %v", vip)
	}
}
