import { StraitGatewayApiClient } from './client'
import { GatewayResource, HttpRouteResource, TransitGatewayResource } from '../types/gateway'

export class GatewaysApi {
  constructor(private client: StraitGatewayApiClient) {}

  public getGateways(namespace?: string, signal?: AbortSignal): Promise<GatewayResource[]> {
    return this.client.get<GatewayResource[]>('/gateways', { namespace }, signal)
  }

  public getTransitGateways(signal?: AbortSignal): Promise<TransitGatewayResource[]> {
    return this.client.get<TransitGatewayResource[]>('/transitgateways', undefined, signal)
  }

  public getHttpRoutes(namespace?: string, signal?: AbortSignal): Promise<HttpRouteResource[]> {
    return this.client.get<HttpRouteResource[]>('/httproutes', { namespace }, signal)
  }
}
