// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package main implements the straitgateway controller manager.
//
// The controller manager runs as a Kubernetes Deployment with leader election.
// It registers all straitgateway controllers and manages their lifecycle.
//
// Leader election ID: straitgateway-controller-leader
package main

import (
	"flag"
	"os"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	sgv1alpha1 "github.com/msaeedb40/straitgateway/api/v1alpha1"
	"github.com/msaeedb40/straitgateway/controllers"
	"github.com/msaeedb40/straitgateway/internal/version"
	polcompiler "github.com/msaeedb40/straitgateway/policy"
	svcmgr "github.com/msaeedb40/straitgateway/service"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(sgv1alpha1.AddToScheme(scheme))
	utilruntime.Must(gwapiv1.Install(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
}

func main() {
	var (
		metricsAddr      string
		probeAddr        string
		leaderElect      bool
		leaderElectionID string
		logLevel         string
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":9090",
		"The address the metrics endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081",
		"The address the health probe endpoint binds to.")
	flag.BoolVar(&leaderElect, "leader-elect", true,
		"Enable leader election for controller manager.")
	flag.StringVar(&leaderElectionID, "leader-election-id", "straitgateway-controller-leader",
		"The name of the resource that leader election will use for holding the leader lock.")
	flag.StringVar(&logLevel, "log-level", "info",
		"Log level: debug, info, warn, error.")
	flag.Parse()

	// Configure structured JSON logger.
	zapLevel := zapcore.InfoLevel
	if logLevel == "debug" {
		zapLevel = zapcore.DebugLevel
	}
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	defer logger.Sync() //nolint:errcheck

	ctrl.SetLogger(zapr.NewLogger(logger))
	setupLog := logger.Named("setup")

	setupLog.Info("initializing straitgateway controller manager",
		zap.String("version", version.Version),
		zap.String("commit", version.Commit),
		zap.String("go", version.GoVersion),
	)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         leaderElect,
		LeaderElectionID:       leaderElectionID,
	})
	if err != nil {
		setupLog.Fatal("unable to create controller manager", zap.Error(err))
	}

	// Shared domain managers
	serviceManager := svcmgr.New(mgr.GetClient(), logger.Named("service-manager"))
	policyCompiler := polcompiler.New(mgr.GetClient(), logger.Named("policy-compiler"))

	// Register all 10 controllers
	if err := (&controllers.ServiceReconciler{
		Client:  mgr.GetClient(),
		Log:     logger.Named("service-controller"),
		Manager: serviceManager,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Fatal("unable to create Service controller", zap.Error(err))
	}

	if err := (&controllers.EndpointSliceReconciler{
		Client:  mgr.GetClient(),
		Log:     logger.Named("endpointslice-controller"),
		Manager: serviceManager,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Fatal("unable to create EndpointSlice controller", zap.Error(err))
	}

	if err := (&controllers.NetworkPolicyReconciler{
		Client:   mgr.GetClient(),
		Log:      logger.Named("networkpolicy-controller"),
		Compiler: policyCompiler,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Fatal("unable to create NetworkPolicy controller", zap.Error(err))
	}

	if err := (&controllers.StraitNetworkPolicyReconciler{
		Client:   mgr.GetClient(),
		Log:      logger.Named("straitnetworkpolicy-controller"),
		Compiler: policyCompiler,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Fatal("unable to create StraitNetworkPolicy controller", zap.Error(err))
	}

	if err := (&controllers.GatewayReconciler{
		Client: mgr.GetClient(),
		Log:    logger.Named("gateway-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Fatal("unable to create Gateway controller", zap.Error(err))
	}

	if err := (&controllers.TransitGatewayReconciler{
		Client: mgr.GetClient(),
		Log:    logger.Named("transit-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Fatal("unable to create Transit controller", zap.Error(err))
	}

	if err := (&controllers.BGPPeerReconciler{
		Client: mgr.GetClient(),
		Log:    logger.Named("bgp-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Fatal("unable to create BGP controller", zap.Error(err))
	}

	if err := (&controllers.IdentityReconciler{
		Client: mgr.GetClient(),
		Log:    logger.Named("identity-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Fatal("unable to create Identity controller", zap.Error(err))
	}

	if err := (&controllers.IPAMReconciler{
		Client: mgr.GetClient(),
		Log:    logger.Named("ipam-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Fatal("unable to create IPAM controller", zap.Error(err))
	}

	if err := (&controllers.NodeNetworkConfigReconciler{
		Client: mgr.GetClient(),
		Log:    logger.Named("nodenetworkconfig-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Fatal("unable to create NodeNetworkConfig controller", zap.Error(err))
	}

	// Health checks.
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Fatal("unable to set up health check", zap.Error(err))
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Fatal("unable to set up ready check", zap.Error(err))
	}

	setupLog.Info("starting straitgateway controller manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Fatal("problem running manager", zap.Error(err))
	}
}
