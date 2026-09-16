import { EbpfHookType } from './node.model';

export type EbpfProgramType =
  | 'cgroup_connect4'
  | 'cgroup_connect6'
  | 'cgroup_sendmsg4'
  | 'cgroup_sendmsg6'
  | 'cgroup_recvmsg4'
  | 'cgroup_recvmsg6'
  | 'xdp'
  | 'tc_ingress'
  | 'tc_egress'
  | 'lsm';

export type EbpfProgramState = 'Loaded' | 'Loading' | 'Failed' | 'Unloaded';
export type EbpfMapType = 'hash' | 'array' | 'lpm_trie' | 'perf_event_array' | 'ringbuf' | 'sk_storage' | 'cgroup_storage';

export interface EbpfProgram {
  readonly id: number;
  readonly name: string;
  readonly type: EbpfProgramType;
  readonly hook: EbpfHookType;
  readonly state: EbpfProgramState;
  readonly nodeName: string;
  readonly tag: string;
  readonly loadedAt?: string;
  readonly error?: string;
  readonly runCount?: number;
  readonly runTimeNs?: number;
}

export interface EbpfMap {
  readonly id: number;
  readonly name: string;
  readonly type: EbpfMapType;
  readonly keySize: number;
  readonly valueSize: number;
  readonly maxEntries: number;
  readonly currentEntries?: number;
  readonly nodeName: string;
}

export interface EbpfAttachment {
  readonly programId: number;
  readonly programName: string;
  readonly hook: EbpfHookType;
  readonly interfaceName?: string;
  readonly cgroupPath?: string;
  readonly priority?: number;
  readonly nodeName: string;
  readonly attachedAt?: string;
}

export interface EbpfNodeSummary {
  readonly nodeName: string;
  readonly programs: EbpfProgram[];
  readonly maps: EbpfMap[];
  readonly attachments: EbpfAttachment[];
  readonly errors: string[];
}

export interface EbpfLoadBalancerEntry {
  readonly frontendIP: string;
  readonly frontendPort: number;
  readonly protocol: string;
  readonly backends: EbpfBackend[];
}

export interface EbpfBackend {
  readonly ip: string;
  readonly port: number;
  readonly weight: number;
  readonly state: 'Active' | 'Draining' | 'Failed';
}
