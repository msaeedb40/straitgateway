export type FlowProtocol = 'TCP' | 'UDP' | 'ICMP' | 'SCTP' | 'Unknown';
export type FlowDirection = 'Ingress' | 'Egress' | 'Transit';
export type FlowState = 'Active' | 'Closing' | 'Closed' | 'Failed';

export interface Flow {
  readonly id: string;
  readonly sourceIP: string;
  readonly destinationIP: string;
  readonly sourcePort: number;
  readonly destinationPort: number;
  readonly protocol: FlowProtocol;
  readonly direction: FlowDirection;
  readonly state: FlowState;
  readonly bytes: number;
  readonly packets: number;
  readonly bytesPerSecond: number;
  readonly packetsPerSecond: number;
  readonly latencyMs: number | null;
  readonly gatewayName: string;
  readonly nodeName: string;
  readonly namespace: string;
  readonly tunnelName: string | null;
  readonly serviceName: string | null;
  readonly startTime: string;
  readonly lastSeenTime: string;
}

export interface FlowFilter {
  namespace?: string;
  cluster?: string;
  nodeName?: string;
  gatewayName?: string;
  protocol?: FlowProtocol;
  sourceIP?: string;
  destinationIP?: string;
  state?: FlowState;
  since?: string;
  until?: string;
  page?: number;
  pageSize?: number;
}

export interface FlowPath {
  readonly flowId: string;
  readonly hops: FlowHop[];
}

export interface FlowHop {
  readonly type: 'Gateway' | 'Tunnel' | 'Node' | 'Service' | 'Endpoint';
  readonly name: string;
  readonly namespace: string;
  readonly latencyMs: number | null;
}
