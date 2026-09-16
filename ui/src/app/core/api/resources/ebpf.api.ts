import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { EbpfNodeSummary, EbpfProgram, EbpfMap, EbpfAttachment, EbpfLoadBalancerEntry } from '../../models/ebpf.model';
import { EbpfHookType } from '../../models/node.model';

@Injectable({ providedIn: 'root' })
export class EbpfApiService {
  private readonly client = inject(ApiClient);

  /** All eBPF state for a specific node */
  getNodeSummary(nodeName: string): Observable<EbpfNodeSummary> {
    return this.client.get<EbpfNodeSummary>(`/v1/nodes/${nodeName}/ebpf`);
  }

  /** All programs across all nodes */
  listPrograms(nodeName?: string): Observable<EbpfProgram[]> {
    return this.client.get<EbpfProgram[]>('/v1/ebpf/programs', {
      params: nodeName ? { nodeName } : undefined,
    });
  }

  /** All maps across all nodes */
  listMaps(nodeName?: string): Observable<EbpfMap[]> {
    return this.client.get<EbpfMap[]>('/v1/ebpf/maps', {
      params: nodeName ? { nodeName } : undefined,
    });
  }

  /** All attachments — cgroup / tcx / xdp / lsm */
  listAttachments(hook?: EbpfHookType, nodeName?: string): Observable<EbpfAttachment[]> {
    return this.client.get<EbpfAttachment[]>('/v1/ebpf/attachments', {
      params: {
        ...(hook ? { hook } : {}),
        ...(nodeName ? { nodeName } : {}),
      },
    });
  }

  /** Load balancer entries managed by eBPF */
  listLoadBalancerEntries(): Observable<EbpfLoadBalancerEntry[]> {
    return this.client.get<EbpfLoadBalancerEntry[]>('/v1/ebpf/lb');
  }
}
