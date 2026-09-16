export type CniPhase = 'Installed' | 'Installing' | 'Failed' | 'NotInstalled' | 'Unknown';

export interface CniConfig {
  readonly cniVersion: string;
  readonly name: string;
  readonly type: string;
  readonly raw: Record<string, unknown>;
}

export interface CniInterface {
  readonly name: string;
  readonly nodeName: string;
  readonly mac: string;
  readonly mtu: number;
  readonly addresses: string[];
}

export interface CniIpAllocation {
  readonly podName: string;
  readonly podNamespace: string;
  readonly nodeName: string;
  readonly ip: string;
  readonly gateway: string;
  readonly subnet: string;
  readonly allocatedAt: string;
}

export interface CniRoute {
  readonly nodeName: string;
  readonly destination: string;
  readonly gateway?: string;
  readonly interface: string;
  readonly metric?: number;
  readonly source: 'Kubernetes' | 'Straitd' | 'CNI' | 'Kernel' | 'Controller';
}

export interface CniNodeStatus {
  readonly nodeName: string;
  readonly phase: CniPhase;
  readonly version?: string;
  readonly configPath?: string;
  readonly podCIDR?: string;
  readonly errors: string[];
}

export interface CniSummary {
  readonly phase: CniPhase;
  readonly config?: CniConfig;
  readonly interfaces: CniInterface[];
  readonly allocations: CniIpAllocation[];
  readonly routes: CniRoute[];
  readonly nodeStatuses: CniNodeStatus[];
  readonly errors: string[];
}
