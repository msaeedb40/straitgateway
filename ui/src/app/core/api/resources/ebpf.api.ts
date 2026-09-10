// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { EbpfMapSummary } from '../api.types';

@Injectable({ providedIn: 'root' })
export class EbpfApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.apiBase(); }

  listMaps(): Observable<EbpfMapSummary[]> {
    return this.http.get<EbpfMapSummary[]>(`${this.base}/api/v1/ebpf/maps`).pipe(
      catchError(() => of(this.getMockMaps()))
    );
  }

  dumpMap(name: string): Observable<any[]> {
    return this.http.get<any[]>(`${this.base}/api/v1/ebpf/maps/${name}/dump`).pipe(
      catchError(() => of([
        { key: '10.244.1.42:48210 -> 10.96.0.10:53', value: { packets: 1240, bytes: 96720, state: 'ESTABLISHED' } },
        { key: '198.51.100.12:53420 -> 10.244.0.15:443', value: { packets: 4209, bytes: 4981200, state: 'ESTABLISHED' } },
      ]))
    );
  }

  private getMockMaps(): EbpfMapSummary[] {
    return [
      {
        id: 1,
        name: 'sg_flow_table',
        type: 'BPF_MAP_TYPE_LRU_HASH',
        maxEntries: 262144,
        currentEntries: 14205,
        flags: 0,
        pinnedPath: '/sys/fs/bpf/straitgateway/sg_flow_table',
        keySize: 48,
        valueSize: 64,
      },
      {
        id: 2,
        name: 'sg_nat_table',
        type: 'BPF_MAP_TYPE_LRU_HASH',
        maxEntries: 131072,
        currentEntries: 8920,
        flags: 0,
        pinnedPath: '/sys/fs/bpf/straitgateway/sg_nat_table',
        keySize: 32,
        valueSize: 48,
      },
      {
        id: 3,
        name: 'sg_maglev_lookup',
        type: 'BPF_MAP_TYPE_ARRAY',
        maxEntries: 65537,
        currentEntries: 65537,
        flags: 0,
        pinnedPath: '/sys/fs/bpf/straitgateway/sg_maglev_lookup',
        keySize: 4,
        valueSize: 4,
      },
      {
        id: 4,
        name: 'sg_policy_rules',
        type: 'BPF_MAP_TYPE_LPM_TRIE',
        maxEntries: 16384,
        currentEntries: 324,
        flags: 1,
        pinnedPath: '/sys/fs/bpf/straitgateway/sg_policy_rules',
        keySize: 8,
        valueSize: 16,
      },
      {
        id: 5,
        name: 'sg_metrics_ring',
        type: 'BPF_MAP_TYPE_RINGBUF',
        maxEntries: 16777216,
        currentEntries: 1048576,
        flags: 0,
        pinnedPath: '/sys/fs/bpf/straitgateway/sg_metrics_ring',
        keySize: 0,
        valueSize: 0,
      },
    ];
  }
}
