# StraitGateway Control Plane Architecture

## Overview
The SG Control Plane is composed of `sg-controller` running in the Kubernetes cluster as a Deployment, reconciling desired state from the Kubernetes API Server and synchronizing it with the node agents (`straitd`).

## Responsibilities
- **Gateway API Watchers**: Watches GatewayClass, Gateway, HTTPRoute, and TCPRoute objects (Gateway API v1.6.1).
- **Service & EndpointSlice Reconciler**: Reconciles ClusterIP, NodePort, and LoadBalancer VIPs into eBPF backend maps without iptables/IPVS.
- **Policy Watcher**: Watches `StraitGatewayPolicy` and standard `NetworkPolicy` objects, driving the CEL policy compilation engine.
- **Node Watcher**: Manages per-node pod CIDRs (`spec.podCIDRs`) for cluster-pool IPAM without RFC1918 assumptions.

## High Availability & Leader Election
`sg-controller` utilizes standard Kubernetes `Lease` locks for active-passive high availability across multiple replicas.
