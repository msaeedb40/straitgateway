import { StraitGatewayApiClient } from './client'
import { BfdSession, BgpPeerSession, RouteEntry, SecurityPolicyRule, WireguardPeer } from '../types/networking'

export class RoutesApi {
  constructor(private client: StraitGatewayApiClient) {}

  public getRoutes(vrfId?: number, signal?: AbortSignal): Promise<RouteEntry[]> {
    return this.client.get<RouteEntry[]>('/routing/routes', { vrfId }, signal)
  }

  public getBgpPeers(signal?: AbortSignal): Promise<BgpPeerSession[]> {
    return this.client.get<BgpPeerSession[]>('/routing/bgp/peers', undefined, signal)
  }

  public getBfdSessions(signal?: AbortSignal): Promise<BfdSession[]> {
    return this.client.get<BfdSession[]>('/routing/bfd/sessions', undefined, signal)
  }

  public getWireguardPeers(signal?: AbortSignal): Promise<WireguardPeer[]> {
    return this.client.get<WireguardPeer[]>('/wireguard/peers', undefined, signal)
  }

  public getSecurityPolicies(namespace?: string, signal?: AbortSignal): Promise<SecurityPolicyRule[]> {
    return this.client.get<SecurityPolicyRule[]>('/security/policies', { namespace }, signal)
  }
}
