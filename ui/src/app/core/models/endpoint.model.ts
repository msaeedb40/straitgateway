import { ObjectMeta, ResourceStatus } from '../api/api.types';
import { Protocol } from './service.model';

export interface EndpointPort {
  readonly name?: string;
  readonly port: number;
  readonly protocol: Protocol;
}

export interface EndpointAddress {
  readonly ip: string;
  readonly hostname?: string;
  readonly nodeName?: string;
  readonly targetRef?: {
    readonly kind: string;
    readonly name: string;
    readonly namespace: string;
  };
}

export interface EndpointSubset {
  readonly addresses: EndpointAddress[];
  readonly notReadyAddresses?: EndpointAddress[];
  readonly ports: EndpointPort[];
}

export interface Endpoint {
  readonly metadata: ObjectMeta;
  readonly subsets: EndpointSubset[];
  readonly status: ResourceStatus;
}
