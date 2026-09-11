// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

export interface NodeStatus {
  name: string;
  ip: string;
  podCIDR: string;
  kernelVersion: string;
  cniReady: boolean;
  serviceReady: boolean;
  policyReady: boolean;
  gatewayReady: boolean;
  kubeProxyReplacement: boolean;
  bpfMapRevision: number;
  wireguardPublicKey?: string;
}

export interface ServiceSummary {
  name: string;
  namespace: string;
  clusterIP: string;
  type: 'ClusterIP' | 'NodePort' | 'LoadBalancer';
  port: number;
  protocol: 'TCP' | 'UDP' | 'SCTP';
  backendCount: number;
  algorithm: string;
}

export interface ListenerSummary {
  name: string;
  protocol: string;
  port: number;
  routeCount: number;
}

export interface GatewaySummary {
  name: string;
  namespace: string;
  gatewayClass: string;
  addresses: string[];
  listeners: ListenerSummary[];
  ready: boolean;
  conditions?: Array<{ type: string; status: string; reason?: string; message?: string }>;
}

export interface GatewayRouteSummary {
  name: string;
  namespace: string;
  kind: 'HTTPRoute' | 'GRPCRoute' | 'TCPRoute' | 'TLSRoute';
  hostnames: string[];
  rulesCount: number;
  parentGateways: string[];
  status: 'Accepted' | 'Programmed' | 'Pending' | 'Error';
}

export interface TunnelPeer {
  clusterID: number;
  tunnelIP: string;
  endpoint: string;
  latencyMs: number;
  txBytes: number;
  rxBytes: number;
  state: 'up' | 'down';
  lastHandshake?: string;
}

export interface BGPPeerStatus {
  peerAddress: string;
  peerASN: number;
  localASN: number;
  sessionState: 'Established' | 'Connect' | 'Active' | 'Idle';
  prefixesAdvertised: number;
  prefixesReceived: number;
  bfdState?: 'Up' | 'Down' | 'AdminDown';
}

export interface KubeEvent {
  namespace: string;
  name: string;
  reason: string;
  message: string;
  type: 'Normal' | 'Warning';
  count: number;
  lastSeen: string;
}

export interface DashboardSummary {
  totalNodes: number;
  readyNodes: number;
  totalGateways: number;
  activeGateways: number;
  totalTunnels: number;
  activeTunnels: number;
  totalServices: number;
  totalFlowsPerSec: number;
  droppedFlowsPerSec: number;
  healthPercentage: number;
}

export interface EbpfMapSummary {
  id: number;
  name: string;
  type: string;
  maxEntries: number;
  currentEntries: number;
  flags: number;
  pinnedPath: string;
  keySize: number;
  valueSize: number;
}

export interface CniNodeStatus {
  nodeName: string;
  podCIDR: string;
  assignedIPs: number;
  totalIPs: number;
  netkitInterface: string;
  mtu: number;
  status: 'Ready' | 'NotReady';
}

export interface EndpointSummary {
  name: string;
  namespace: string;
  serviceName: string;
  addressType: string;
  readyEndpoints: Array<{ ip: string; nodeName?: string; port: number }>;
  notReadyEndpoints?: Array<{ ip: string; nodeName?: string; port: number }>;
}

export interface StraitNetworkPolicy {
  name: string;
  namespace: string;
  policyType: 'Ingress' | 'Egress' | 'Both';
  rulesCount: number;
  appliedPodsCount: number;
  enforcementMode: 'eBPF' | 'Enforcing' | 'Audit';
}
