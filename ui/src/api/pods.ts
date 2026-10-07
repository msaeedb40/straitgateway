import { StraitGatewayApiClient } from './client'
import { PodInfo } from '../types/cluster'

export class PodsApi {
  constructor(private client: StraitGatewayApiClient) {}

  public getPods(namespace?: string, signal?: AbortSignal): Promise<PodInfo[]> {
    return this.client.get<PodInfo[]>('/pods', { namespace }, signal)
  }
}
