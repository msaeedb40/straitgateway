import { StraitGatewayApiClient } from './client'
import { NodeInfo } from '../types/cluster'

export class NodesApi {
  constructor(private client: StraitGatewayApiClient) {}

  public getNodes(signal?: AbortSignal): Promise<NodeInfo[]> {
    return this.client.get<NodeInfo[]>('/nodes', undefined, signal)
  }

  public getNodeByName(name: string, signal?: AbortSignal): Promise<NodeInfo> {
    return this.client.get<NodeInfo>(`/nodes/${encodeURIComponent(name)}`, undefined, signal)
  }
}
