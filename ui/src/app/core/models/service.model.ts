import { ObjectMeta, ResourceStatus } from '../api/api.types';

export type ServiceType = 'ClusterIP' | 'NodePort' | 'LoadBalancer' | 'ExternalName';
export type Protocol = 'TCP' | 'UDP' | 'SCTP';

export interface ServicePort {
  readonly name?: string;
  readonly protocol: Protocol;
  readonly port: number;
  readonly targetPort: number | string;
  readonly nodePort?: number;
}

export interface ServiceSpec {
  readonly type: ServiceType;
  readonly clusterIP: string;
  readonly clusterIPs: string[];
  readonly ports: ServicePort[];
  readonly selector?: Record<string, string>;
  readonly externalName?: string;
}

export interface KubernetesService {
  readonly metadata: ObjectMeta;
  readonly spec: ServiceSpec;
  readonly status: ResourceStatus;
}
