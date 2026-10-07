import { StraitGatewayApiClient } from './client'
import { ServiceInfo } from '../types/cluster'

export class ServicesApi {
  constructor(private client: StraitGatewayApiClient) {}

  public getServices(namespace?: string, signal?: AbortSignal): Promise<ServiceInfo[]> {
    return this.client.get<ServiceInfo[]>('/services', { namespace }, signal)
  }
}
