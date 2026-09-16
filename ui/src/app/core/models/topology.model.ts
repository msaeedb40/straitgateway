export type TopologyNodeKind =
  | 'Cluster'
  | 'Gateway'
  | 'Node'
  | 'CNI'
  | 'Pod'
  | 'Service'
  | 'Endpoint'
  | 'Tunnel';

export type TopologyEdgeKind =
  | 'GatewayToNode'
  | 'GatewayToTunnel'
  | 'TunnelToGateway'
  | 'NodeToService'
  | 'ServiceToEndpoint'
  | 'NodeToCNI';

export type HealthStatus = 'Healthy' | 'Degraded' | 'Failed' | 'Unknown';

export interface TopologyNode {
  readonly id: string;
  readonly kind: TopologyNodeKind;
  readonly name: string;
  readonly namespace: string;
  readonly cluster: string;
  readonly health: HealthStatus;
  readonly labels?: Record<string, string>;
}

export interface TopologyEdge {
  readonly id: string;
  readonly kind: TopologyEdgeKind;
  readonly sourceId: string;
  readonly targetId: string;
  /** Bytes/sec derived from Prometheus — null when metrics unavailable */
  readonly trafficBytesPerSec: number | null;
  /** Packets/sec derived from Prometheus — null when metrics unavailable */
  readonly trafficPacketsPerSec: number | null;
  readonly errorRate: number | null;
  readonly latencyMs: number | null;
}

export interface TopologyGraph {
  readonly nodes: TopologyNode[];
  readonly edges: TopologyEdge[];
  readonly generatedAt: string;
  readonly namespace: string;
  readonly cluster: string;
}
