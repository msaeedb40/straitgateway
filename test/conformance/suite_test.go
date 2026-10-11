// Package conformance provides Sonobuoy-compatible CNI conformance tests and
// Gateway API conformance tests for StraitGateway.
//
// Run with:
//
//	go test -v -count=1 -tags conformance ./test/conformance/... \
//	  -gateway-class straitgateway \
//	  -kubeconfig $HOME/.kube/config
package conformance_test

import (
	"flag"
	"io/fs"
	"os"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance"
	confv1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"

	// Register our API scheme
	_ "github.com/straitgateway/straitgateway/test/conformance/internal"
)

var (
	gatewayClassName = flag.String("gateway-class", "straitgateway", "GatewayClass name to test")
	kubeconfig       = flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "Path to kubeconfig")
	reportOutput     = flag.String("report-output", "", "Path to write conformance report JSON")
	cleanupOnFailure = flag.Bool("cleanup-on-failure", false, "Clean up resources even when tests fail")
	allNamespaces    = flag.Bool("all-namespaces", false, "Run tests across all namespaces")
)

// TestGatewayAPIConformance is the upstream Gateway API conformance suite entry point.
// Sonobuoy will discover and invoke this test via the standard Go test binary mechanism.
func TestGatewayAPIConformance(t *testing.T) {
	flag.Parse()

	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("failed to get kubeconfig: %v", err)
	}

	c, err := client.New(cfg, client.Options{})
	if err != nil {
		t.Fatalf("failed to create k8s client: %v", err)
	}

	// Register Gateway API CRDs into the scheme
	if err := gatewayv1.Install(c.Scheme()); err != nil {
		t.Fatalf("failed to install Gateway API scheme: %v", err)
	}

	// StraitGateway supported features (Phase 10/11 capabilities).
	// In gateway-api v1.6+, SupportedFeatures is []features.FeatureName.
	supportedFeatures := []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteMethodMatching,
		features.SupportHTTPRouteQueryParamMatching,
		features.SupportHTTPRouteResponseHeaderModification,
		// SupportHTTPRouteRequestHeaderModification was renamed in v1.6:
		features.SupportHTTPRouteBackendRequestHeaderModification,
		// SupportHTTPRouteRequestRedirect was split in v1.6:
		features.SupportHTTPRoutePortRedirect,
		features.SupportHTTPRouteSchemeRedirect,
		features.SupportHTTPRoutePathRedirect,
		// SupportHTTPRouteURLRewrite was split in v1.6:
		features.SupportHTTPRouteHostRewrite,
		features.SupportHTTPRoutePathRewrite,
		features.SupportHTTPRouteDestinationPortMatching,
		features.SupportGRPCRoute,
		features.SupportTLSRoute,
		features.SupportReferenceGrant,
	}

	// ExemptFeatures: kernel-privilege tests that require privileged pods.
	// In v1.6+, this is also []features.FeatureName (not sets.Set).
	exemptFeatures := []features.FeatureName{}

	opts := suite.ConformanceOptions{
		Client:               c,
		BaseManifests:        "manifests/base.yaml",
		MeshManifests:        "",
		ManifestFS:           []fs.FS{&conformance.Manifests},
		ConfigurableOptions: suite.ConfigurableOptions{
			GatewayClassName:           *gatewayClassName,
			CleanupBaseResources:       true,
			SupportedFeatures:          supportedFeatures,
			ExemptFeatures:             exemptFeatures,
			EnableAllSupportedFeatures: false,
			RunTest:                    "",
			AllowCRDsMismatch:          false,
			// GatewayConformanceProfileName was renamed to GatewayHTTPConformanceProfileName in v1.6
			ConformanceProfiles: []suite.ConformanceProfileName{
				suite.GatewayHTTPConformanceProfileName,
			},
			Implementation: confv1.Implementation{
				Organization: "straitgateway",
				Project:      "straitgateway",
				URL:          "https://github.com/straitgateway/straitgateway",
				Version:      "v1.0",
				Contact:      []string{"@straitgateway/maintainers"},
			},
			ReportOutputPath: *reportOutput,
		},
	}

	_ = cleanupOnFailure
	_ = allNamespaces
	_ = kubeconfig

	cSuite, err := suite.NewConformanceTestSuite(opts)
	if err != nil {
		t.Fatalf("failed to create conformance suite: %v", err)
	}

	cSuite.Setup(t, tests.ConformanceTests)
	if err = cSuite.Run(t, tests.ConformanceTests); err != nil {
		t.Fatalf("conformance suite run error: %v", err)
	}

	report, err := cSuite.Report()
	if err != nil {
		t.Fatalf("failed to generate conformance report: %v", err)
	}

	if *reportOutput != "" {
		if err := os.WriteFile(*reportOutput, []byte(report.String()), 0o644); err != nil {
			t.Fatalf("failed to write report: %v", err)
		}
		t.Logf("conformance report written to %s", *reportOutput)
	}
}
