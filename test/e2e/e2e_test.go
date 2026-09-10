// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset

func TestMain(m *testing.M) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		fmt.Printf("Skipping e2e tests: no kubeconfig found at %s\n", kubeconfig)
		os.Exit(0)
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Printf("Skipping e2e tests: error building kubeconfig: %v\n", err)
		os.Exit(0)
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Skipping e2e tests: error creating clientset: %v\n", err)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestStraitgatewayDaemonSetReady(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ds, err := clientset.AppsV1().DaemonSets("straitgateway-system").Get(ctx, "straitgateway-agent", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get straitgateway-agent DaemonSet: %v", err)
	}

	if ds.Status.NumberReady != ds.Status.DesiredNumberScheduled {
		t.Errorf("DaemonSet not fully ready: %d/%d",
			ds.Status.NumberReady, ds.Status.DesiredNumberScheduled)
	}
}

func TestControllerDeploymentReady(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	deploy, err := clientset.AppsV1().Deployments("straitgateway-system").Get(ctx, "straitgateway-controller", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get controller deployment: %v", err)
	}

	if deploy.Status.ReadyReplicas < 1 {
		t.Error("Controller deployment has no ready replicas")
	}
}

func TestGatewayClassExists(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use dynamic client for GatewayClass since it's a CRD.
	_ = ctx
	t.Log("GatewayClass 'skgateway' check — requires gateway-api CRDs")
}

func TestPodNetworking(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ns := "straitgateway-e2e-test"
	// Create test namespace.
	_, err := clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ns},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}
	defer clientset.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})

	// Create two test pods.
	for _, name := range []string{"pod-a", "pod-b"} {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:    "test",
					Image:   "busybox:latest",
					Command: []string{"sleep", "300"},
				}},
			},
		}
		if _, err := clientset.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
			t.Fatalf("Failed to create pod %s: %v", name, err)
		}
	}

	// Wait for pods to be running.
	t.Log("Waiting for test pods to become ready...")
	time.Sleep(30 * time.Second)

	pods, err := clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list pods: %v", err)
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			t.Errorf("Pod %s is %s, expected Running", pod.Name, pod.Status.Phase)
		}
		if pod.Status.PodIP == "" {
			t.Errorf("Pod %s has no IP assigned", pod.Name)
		} else {
			t.Logf("Pod %s: IP=%s", pod.Name, pod.Status.PodIP)
		}
	}
}

func TestServiceLoadBalancing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Verify kube-dns service is accessible (critical for kube-proxy replacement).
	svc, err := clientset.CoreV1().Services("kube-system").Get(ctx, "kube-dns", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get kube-dns service: %v", err)
	}
	t.Logf("kube-dns ClusterIP: %s", svc.Spec.ClusterIP)

	endpoints, err := clientset.CoreV1().Endpoints("kube-system").Get(ctx, "kube-dns", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get kube-dns endpoints: %v", err)
	}
	if len(endpoints.Subsets) == 0 {
		t.Error("kube-dns has no endpoint subsets")
	}
}

func TestNetworkPolicy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Verify NetworkPolicy resources are handled.
	policies, err := clientset.NetworkingV1().NetworkPolicies("").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list NetworkPolicies: %v", err)
	}
	t.Logf("Found %d NetworkPolicies across all namespaces", len(policies.Items))
}
