// Package internal registers the StraitGateway runtime scheme so conformance
// test files can import it as a blank import side-effect.
package internal

import (
	"k8s.io/client-go/kubernetes/scheme"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func init() {
	_ = gatewayv1.Install(scheme.Scheme)
	_ = gatewayv1alpha2.Install(scheme.Scheme)
	_ = gatewayv1beta1.Install(scheme.Scheme)
}
