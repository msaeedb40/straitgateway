export interface GatewayResource {
  name: string
  namespace: string
  gatewayClassName: string
  listeners: Array<{
    name: string
    protocol: 'HTTP' | 'HTTPS' | 'TCP' | 'UDP' | 'TLS'
    port: number
    routesCount: number
  }>
  addresses: string[]
  programmed: boolean
  ready: boolean
}

export interface TransitGatewayResource {
  name: string
  namespace: string
  asn: number
  routerId: string
  vrfs: Array<{
    name: string
    id: number
    routeTableId: number
    cidrs: string[]
  }>
  peersCount: number
  tunnelsCount: number
  status: 'Active' | 'Configuring' | 'Failed'
}

export interface HttpRouteResource {
  name: string
  namespace: string
  parentRefs: string[]
  hostnames: string[]
  rulesCount: number
}
