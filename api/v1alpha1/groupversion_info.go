// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupVersion is group version used to register these objects.
var GroupVersion = schema.GroupVersion{Group: "straitgateway.io", Version: "v1alpha1"}

// SchemeBuilder is used to add functions to this group's scheme.
var SchemeBuilder = &runtime.SchemeBuilder{localSchemeBuilder.AddToScheme}

// AddToScheme adds the types in this group-version to the given scheme.
var AddToScheme = SchemeBuilder.AddToScheme

var localSchemeBuilder = runtime.SchemeBuilder{
	func(s *runtime.Scheme) error {
		s.AddKnownTypes(GroupVersion,
			&StraitNetworkPolicy{},
			&StraitNetworkPolicyList{},
			&TransitGateway{},
			&TransitGatewayList{},
			&TransitSegment{},
			&TransitSegmentList{},
			&TransitSegmentAttachment{},
			&TransitSegmentAttachmentList{},
			&TransitSegmentRoute{},
			&TransitSegmentRouteList{},
			&BGPPeer{},
			&BGPPeerList{},
			&BFDSession{},
			&BFDSessionList{},
			&NodeNetworkConfig{},
			&NodeNetworkConfigList{},
			&ClusterNetworkConfig{},
			&ClusterNetworkConfigList{},
		)
		metav1.AddToGroupVersion(s, GroupVersion)
		return nil
	},
}
