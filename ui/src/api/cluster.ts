import { StraitGatewayApiClient } from './client'
import { ClusterOverview } from '../types/cluster'

export class ClusterApi {
  constructor(private client: StraitGatewayApiClient) {}

  public getOverview(signal?: AbortSignal): Promise<ClusterOverview> {
    return this.client.get<ClusterOverview>('/cluster/overview', undefined, signal)
  }

  public getHealth(signal?: AbortSignal): Promise<{ status: string; uptime: number }> {
    return this.client.get<{ status: string; uptime: number }>('/cluster/health', undefined, signal)
  }
}
