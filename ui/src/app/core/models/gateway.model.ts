import { ObjectMeta, ResourceStatus, ReconciliationPhase } from '../api/api.types';

export type GatewayType = 'Ingress' | 'Egress' | 'Transit';

export interface GatewayAddress {
  readonly type: 'IPAddress' | 'Hostname';
  readonly value: string;
}

export interface GatewayListener {
  readonly name: string;
  readonly protocol: string;
  readonly port: number;
  readonly allowedRoutes?: string;
}

export interface GatewaySpec {
  readonly type: GatewayType;
  readonly listeners: GatewayListener[];
  readonly addresses?: GatewayAddress[];
  readonly nodeSelector?: Record<string, string>;
}

export interface GatewayStatus extends ResourceStatus {
  readonly addresses: GatewayAddress[];
  readonly listeners: GatewayListenerStatus[];
  readonly phase: ReconciliationPhase;
  readonly routeCount: number;
  readonly tunnelCount: number;
  readonly activeFlowCount: number;
}

export interface GatewayListenerStatus {
  readonly name: string;
  readonly attachedRoutes: number;
  readonly conditions: import('../api/api.types').Condition[];
}

export interface Gateway {
  readonly metadata: ObjectMeta;
  readonly spec: GatewaySpec;
  readonly status: GatewayStatus;
}

export interface GatewayCreateRequest {
  readonly name: string;
  readonly namespace: string;
  readonly type: GatewayType;
  readonly listeners: GatewayListener[];
  readonly nodeSelector?: Record<string, string>;
  readonly labels?: Record<string, string>;
  readonly annotations?: Record<string, string>;
}

export interface GatewayUpdateRequest {
  readonly spec: Partial<GatewaySpec>;
}

export interface GatewayDeleteImpact {
  readonly routeCount: number;
  readonly tunnelCount: number;
  readonly activeFlowCount: number;
}
