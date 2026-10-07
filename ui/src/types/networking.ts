export interface RouteEntry {
  destinationCIDR: string
  prefixLen: number
  nextHop: string
  outputInterface: string
  routeType: 'Local' | 'Cluster' | 'BGP' | 'WireGuard' | 'Static'
  vrfId: number
  metric: number
  status: 'Active' | 'Withdrawn'
}

export interface BgpPeerSession {
  peerAddress: string
  asn: number
  state: 'Established' | 'OpenConfirm' | 'Active' | 'Connect' | 'Idle'
  uptimeSeconds: number
  prefixesReceived: number
  prefixesAdvertised: number
  holdTime: number
  keepaliveTime: number
}

export interface BfdSession {
  remoteAddress: string
  state: 'Up' | 'Down' | 'AdminDown' | 'Init'
  detectMult: number
  txIntervalMs: number
  rxIntervalMs: number
  lastTransition: string
  fastFailoverTriggered: boolean
}

export interface WireguardPeer {
  publicKey: string
  endpoint: string
  allowedIPs: string[]
  latestHandshake: string
  transferRxBytes: number
  transferTxBytes: number
  persistentKeepalive: number
}

export interface SecurityPolicyRule {
  id: string
  name: string
  namespace: string
  action: 'Allow' | 'Deny'
  direction: 'Ingress' | 'Egress'
  sourceSelector: string
  destinationSelector: string
  ports: string[]
  protocol: string
  hitsCount: number
}
