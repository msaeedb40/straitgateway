import { StraitGatewayApiClient } from './client'
import { ClusterApi } from './cluster'
import { NodesApi } from './nodes'
import { PodsApi } from './pods'
import { ServicesApi } from './services'
import { GatewaysApi } from './gateways'
import { RoutesApi } from './routes'
import { ObservabilityApi } from './observability'
import { getRuntimeConfig } from '../config/runtime'

export class StraitGatewayApiHub {
  public readonly client: StraitGatewayApiClient
  public readonly cluster: ClusterApi
  public readonly nodes: NodesApi
  public readonly pods: PodsApi
  public readonly services: ServicesApi
  public readonly gateways: GatewaysApi
  public readonly routes: RoutesApi
  public readonly observability: ObservabilityApi

  constructor(baseUrl?: string) {
    const effectiveBaseUrl = baseUrl || getRuntimeConfig().apiBaseUrl
    this.client = new StraitGatewayApiClient(effectiveBaseUrl)
    this.cluster = new ClusterApi(this.client)
    this.nodes = new NodesApi(this.client)
    this.pods = new PodsApi(this.client)
    this.services = new ServicesApi(this.client)
    this.gateways = new GatewaysApi(this.client)
    this.routes = new RoutesApi(this.client)
    this.observability = new ObservabilityApi(this.client)
  }
}

// Default singleton instance using dynamic runtime parameters
export const api = new StraitGatewayApiHub()
