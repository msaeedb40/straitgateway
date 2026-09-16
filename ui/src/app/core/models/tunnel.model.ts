import { ObjectMeta, ResourceStatus } from '../api/api.types';

export type TunnelType = 'WireGuard' | 'IPSec' | 'VXLAN' | 'Geneve';
export type TunnelPhase = 'Pending' | 'Connecting' | 'Connected' | 'Degraded' | 'Failed' | 'Unknown';

export interface TunnelEndpoint {
  readonly address: string;
  readonly port: number;
  readonly publicKey?: string;
}

export interface TunnelSpec {
  readonly type: TunnelType;
  readonly localGateway: string;
  readonly remoteGateway: string;
  readonly localEndpoint: TunnelEndpoint;
  readonly remoteEndpoint: TunnelEndpoint;
  readonly allowedCIDRs: string[];
  readonly mtu?: number;
}

export interface TunnelStatus extends Omit<ResourceStatus, 'phase'> {
  readonly phase: TunnelPhase;
  readonly lastHandshake?: string;
  readonly rxBytes?: number;
  readonly txBytes?: number;
  readonly latencyMs?: number;
  readonly activeFlowCount?: number;
}

export interface Tunnel {
  readonly metadata: ObjectMeta;
  readonly spec: TunnelSpec;
  readonly status: TunnelStatus;
}

export interface TunnelCreateRequest {
  readonly name: string;
  readonly namespace: string;
  readonly type: TunnelType;
  readonly localGateway: string;
  readonly remoteGateway: string;
  readonly localEndpoint: TunnelEndpoint;
  readonly remoteEndpoint: TunnelEndpoint;
  readonly allowedCIDRs: string[];
  readonly mtu?: number;
}

export interface TunnelUpdateRequest {
  readonly spec: Partial<TunnelSpec>;
}
