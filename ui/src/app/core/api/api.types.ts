// ─── Generic response wrappers ───────────────────────────────────────────────

export interface ApiResponse<T> {
  readonly data: T;
  readonly meta?: ResponseMeta;
}

export interface PaginatedResponse<T> {
  readonly items: T[];
  readonly meta: PaginationMeta;
}

export interface ResponseMeta {
  readonly resourceVersion?: string;
  readonly timestamp?: string;
}

export interface PaginationMeta {
  readonly total: number;
  readonly page: number;
  readonly pageSize: number;
  readonly continue?: string;
}

export interface ListParams {
  namespace?: string;
  cluster?: string;
  page?: number;
  pageSize?: number;
  labelSelector?: string;
  fieldSelector?: string;
  continue?: string;
}

// ─── Resource lifecycle ───────────────────────────────────────────────────────

export type ReconciliationPhase =
  | 'Pending'
  | 'Reconciling'
  | 'Ready'
  | 'Failed'
  | 'Unknown';

export type ConditionStatus = 'True' | 'False' | 'Unknown';

export interface Condition {
  readonly type: string;
  readonly status: ConditionStatus;
  readonly reason: string;
  readonly message: string;
  readonly lastTransitionTime: string;
}

export interface ResourceStatus {
  readonly phase: ReconciliationPhase;
  readonly conditions: Condition[];
  readonly observedGeneration?: number;
  readonly message?: string;
}

// ─── Kubernetes metadata ──────────────────────────────────────────────────────

export interface ObjectMeta {
  readonly name: string;
  readonly namespace: string;
  readonly uid: string;
  readonly resourceVersion: string;
  readonly generation: number;
  readonly creationTimestamp: string;
  readonly labels?: Record<string, string>;
  readonly annotations?: Record<string, string>;
}

// ─── Connection health ────────────────────────────────────────────────────────

export type ConnectionStatus = 'connected' | 'unavailable' | 'checking';

export interface BackendHealth {
  readonly controller: ConnectionStatus;
  readonly prometheus: ConnectionStatus;
  readonly grafana: ConnectionStatus;
  readonly jaeger: ConnectionStatus;
  readonly logs: ConnectionStatus;
}
