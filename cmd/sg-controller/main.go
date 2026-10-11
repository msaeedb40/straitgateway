package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
	"github.com/straitgateway/straitgateway/controllers/endpoint"
	"github.com/straitgateway/straitgateway/controllers/gateway"
	"github.com/straitgateway/straitgateway/controllers/node"
	"github.com/straitgateway/straitgateway/controllers/peering"
	"github.com/straitgateway/straitgateway/controllers/route"
	"github.com/straitgateway/straitgateway/controllers/security"
	"github.com/straitgateway/straitgateway/controllers/service"
	"github.com/straitgateway/straitgateway/controllers/transitgateway"
	internalgw "github.com/straitgateway/straitgateway/internal/gateway"
	"github.com/straitgateway/straitgateway/internal/loadbalancer"
	"github.com/straitgateway/straitgateway/internal/mesh"
	"github.com/straitgateway/straitgateway/internal/mesh/clustermesh"
	"github.com/straitgateway/straitgateway/internal/routing"
	"github.com/straitgateway/straitgateway/internal/routing/bgp"
	"github.com/straitgateway/straitgateway/internal/wireguard"
	"github.com/straitgateway/straitgateway/pkg/version"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(discoveryv1.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(gatewayv1.Install(scheme))
}

func main() {
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	setupLog.Info("Starting sg-controller", "version", version.Info())

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "straitgateway-controller-leader",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Internal subsystems
	fib := routing.NewFIBEngine()
	routeMgr := routing.NewRouteManager(fib, nil)
	bgpSpeaker := bgp.NewSpeaker(65001, net.ParseIP("127.0.0.1"), routeMgr)
	wgMgr, _ := wireguard.NewManager("wg0", 51820)
	lbMgr := loadbalancer.NewServiceManager(nil)
	identEngine := mesh.NewIdentityEngine()
	policyCompiler := mesh.NewPolicyCompiler(identEngine, nil)
	nodeProxy := mesh.NewNodeProxy()

	// 1. TransitGateway Controller
	if err = (&transitgateway.TransitGatewayReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		FIB:    fib,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create TransitGateway controller")
		os.Exit(1)
	}

	// 2. GatewayPeer Controller
	if err = (&peering.GatewayPeerReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		BGP:       bgpSpeaker,
		WireGuard: wgMgr,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create GatewayPeer controller")
		os.Exit(1)
	}

	// 3. TransitRoute Controller
	if err = (&route.TransitRouteReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Router: routeMgr,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create TransitRoute controller")
		os.Exit(1)
	}

	// 4. Gateway API Controllers (Gateway & HTTPRoute)
	if err = (&gateway.GatewayReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Proxy:  nodeProxy,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Gateway controller")
		os.Exit(1)
	}

	httpRouteEngine := internalgw.NewHTTPRouteEngine()
	if err = (&gateway.HTTPRouteReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Engine: httpRouteEngine,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create HTTPRoute controller")
		os.Exit(1)
	}

	// 5. ClusterMesh Controllers
	cmMgr, _ := clustermesh.NewManager(clustermesh.DefaultConfig(), wgMgr, routeMgr)
	if err = (&peering.ClusterPeeringReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		ClusterMeshMgr: cmMgr,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create ClusterPeering controller")
		os.Exit(1)
	}
	if err = (&peering.GlobalServiceReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		ClusterMeshMgr: cmMgr,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create GlobalService controller")
		os.Exit(1)
	}

	// 5. SecurityPolicy Controller
	if err = (&security.SecurityPolicyReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Compiler: policyCompiler,
		Identity: identEngine,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create SecurityPolicy controller")
		os.Exit(1)
	}

	// 6. Service & Endpoint Controllers
	if err = (&service.ServiceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		LB:     lbMgr,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Service controller")
		os.Exit(1)
	}
	if err = (&endpoint.EndpointSliceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		LB:     lbMgr,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create EndpointSlice controller")
		os.Exit(1)
	}

	// 7. Node Topology Controller
	if err = (&node.NodeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Router: routeMgr,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Node controller")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("Starting StraitGateway Manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
	fmt.Println("Exiting sg-controller")
}
