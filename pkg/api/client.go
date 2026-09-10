// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package api provides Kubernetes client factory and CIDR discovery utilities.
package api

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClientset creates a Kubernetes clientset.
// Uses in-cluster config if available, otherwise falls back to kubeconfig.
func NewClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig for development.
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, fmt.Errorf("creating k8s config: %w", err)
		}
	}
	return kubernetes.NewForConfig(config)
}

// DiscoverPodCIDR reads the podCIDR from node.spec.podCIDR for the given node.
func DiscoverPodCIDR(ctx context.Context, client *kubernetes.Clientset, nodeName string) (string, error) {
	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("getting node %s: %w", nodeName, err)
	}
	if node.Spec.PodCIDR == "" {
		return "", fmt.Errorf("node %s has no podCIDR assigned", nodeName)
	}
	return node.Spec.PodCIDR, nil
}

// DiscoverServiceCIDR attempts to discover the service CIDR by inspecting
// the kubernetes default service ClusterIP.
func DiscoverServiceCIDR(ctx context.Context, client *kubernetes.Clientset) (string, error) {
	svc, err := client.CoreV1().Services("default").Get(ctx, "kubernetes", metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("getting kubernetes service: %w", err)
	}
	// The kubernetes service ClusterIP is typically x.x.x.1 in the service CIDR.
	return svc.Spec.ClusterIP + "/12", nil // approximate; actual CIDR from API server flags
}

// GetNodeInternalIP returns the InternalIP of a Kubernetes node.
func GetNodeInternalIP(node *corev1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}
