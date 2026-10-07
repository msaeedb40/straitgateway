export interface DataplaneMetrics {
  packetsTotal: number
  bytesTotal: number
  dropsTotal: number
  policyDeniedTotal: number
  routeLookupTotal: number
  routeMissTotal: number
  conntrackEntries: number
  natTranslations: number
  bgpSessionsActive: number
  bfdSessionsActive: number
  wireguardPeersCount: number
  wireguardRxBytes: number
  wireguardTxBytes: number
  serviceLbRequests: number
  timestamp: string
}

export interface FlowLogEntry {
  id: string
  timestamp: string
  sourceIp: string
  destinationIp: string
  sourcePort: number
  destinationPort: number
  protocol: 'TCP' | 'UDP' | 'ICMP'
  sourceIdentity: number
  destinationIdentity: number
  action: 'ALLOW' | 'DENY' | 'REDIRECT'
  node: string
  interface: string
  bytes: number
  packets: number
  ruleName?: string
}

export interface EbpfProgramStatus {
  id: number
  name: string
  type: 'XDP' | 'TCX_INGRESS' | 'TCX_EGRESS' | 'CGROUP_SOCK_ADDR' | 'NETKIT'
  device: string
  instructions: number
  runTimeTotalNs: number
  runCount: number
  verified: boolean
}
