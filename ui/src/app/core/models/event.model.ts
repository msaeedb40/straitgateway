export type EventSeverity = 'Normal' | 'Warning' | 'Error';
export type EventResourceKind =
  | 'Gateway'
  | 'Node'
  | 'Tunnel'
  | 'Flow'
  | 'Service'
  | 'Endpoint'
  | 'EbpfProgram'
  | 'CniConfig'
  | 'Configuration';

export interface StraitEvent {
  readonly uid: string;
  readonly timestamp: string;
  readonly severity: EventSeverity;
  readonly resourceKind: EventResourceKind;
  readonly resourceName: string;
  readonly namespace: string;
  readonly cluster: string;
  readonly component: string;
  readonly reason: string;
  readonly message: string;
  readonly count: number;
  readonly firstTime: string;
  readonly lastTime: string;
}

export interface EventFilter {
  namespace?: string;
  cluster?: string;
  severity?: EventSeverity;
  resourceKind?: EventResourceKind;
  resourceName?: string;
  component?: string;
  search?: string;
  since?: string;
  until?: string;
  page?: number;
  pageSize?: number;
}
