export interface ClusterOverview {
  name: string
  version: string
  kubernetesVersion: string
  kernelVersion: string
  ebpfVersion: string
  status: 'Healthy' | 'Degraded' | 'Critical'
  nodesCount: number
  podsCount: number
  servicesCount: number
  gatewaysCount: number
  activeWireguardPeers: number
  bgpSessionsUp: number
  uptimeSeconds: number
}

export interface NodeInfo {
  name: string
  ip: string
  podCIDR: string
  role: 'control-plane' | 'worker' | 'transit-node'
  status: 'Ready' | 'NotReady'
  kernelVersion: string
  netkitSupported: boolean
  tcxAttached: boolean
  ebpfLoadedCount: number
  wireguardPubKey: string
  wireguardEndpoint: string
  cpuUsagePercent: number
  memoryUsagePercent: number
}

export interface PodInfo {
  namespace: string
  name: string
  ip: string
  nodeName: string
  netkitInterface: string
  status: 'Running' | 'Pending' | 'Failed' | 'Succeeded'
  securityIdentity: number
  policyEnforced: boolean
  createdAt: string
}

export interface ServiceInfo {
  namespace: string
  name: string
  type: 'ClusterIP' | 'NodePort' | 'LoadBalancer'
  clusterIP: string
  ports: Array<{
    name: string
    protocol: 'TCP' | 'UDP'
    port: number
    targetPort: number | string
  }>
  endpointsCount: number
  algorithm: 'maglev' | 'round-robin' | 'random'
}
