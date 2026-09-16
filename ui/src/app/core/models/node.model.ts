import { ObjectMeta, ResourceStatus, Condition } from '../api/api.types';

export interface NodeAddress {
  readonly type: 'InternalIP' | 'ExternalIP' | 'Hostname';
  readonly address: string;
}

export interface NodeSystemInfo {
  readonly kernelVersion: string;
  readonly osImage: string;
  readonly containerRuntimeVersion: string;
  readonly kubeletVersion: string;
  readonly architecture: string;
}

export interface NodeCapacity {
  readonly cpu: string;
  readonly memory: string;
  readonly pods: string;
}

export interface NodeSpec {
  readonly podCIDR: string;
  readonly podCIDRs: string[];
  readonly providerID?: string;
  readonly taints?: NodeTaint[];
}

export interface NodeTaint {
  readonly key: string;
  readonly value?: string;
  readonly effect: 'NoSchedule' | 'PreferNoSchedule' | 'NoExecute';
}

// ─── Kubernetes-layer state ───────────────────────────────────────────────────
export interface NodeKubernetesStatus {
  readonly conditions: Condition[];
  readonly addresses: NodeAddress[];
  readonly capacity: NodeCapacity;
  readonly allocatable: NodeCapacity;
  readonly systemInfo: NodeSystemInfo;
  readonly ready: boolean;
}

// ─── StraitGateway-layer state ────────────────────────────────────────────────
export interface StraitdStatus {
  readonly running: boolean;
  readonly version: string;
  readonly uptime?: string;
  readonly socketPath: string;
  readonly lastHeartbeat?: string;
}

// ─── eBPF-layer state ─────────────────────────────────────────────────────────
export type EbpfHookType = 'cgroup' | 'tcx' | 'xdp' | 'lsm';

export interface NodeEbpfStatus {
  readonly loaded: boolean;
  readonly programCount: number;
  readonly mapCount: number;
  readonly errors: string[];
}

// ─── CNI-layer state ──────────────────────────────────────────────────────────
export interface NodeCniStatus {
  readonly installed: boolean;
  readonly version?: string;
  readonly configPath?: string;
  readonly errors: string[];
}

// ─── Composite node ───────────────────────────────────────────────────────────
export interface Node {
  readonly metadata: ObjectMeta;
  readonly spec: NodeSpec;
  readonly kubernetes: NodeKubernetesStatus;
  readonly straitd: StraitdStatus;
  readonly ebpf: NodeEbpfStatus;
  readonly cni: NodeCniStatus;
  readonly status: ResourceStatus;
}
