// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { PacketCapture } from '../api.types';

@Injectable({ providedIn: 'root' })
export class PacketApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.apiBase(); }

  list(limit = 200, action?: string): Observable<PacketCapture[]> {
    const params = new URLSearchParams();
    params.set('limit', limit.toString());
    if (action) params.set('action', action);

    return this.http.get<PacketCapture[]>(`${this.base}/api/v1/packets?${params.toString()}`).pipe(
      catchError(() => of(this.getMockPackets(action)))
    );
  }

  private getMockPackets(action?: string): PacketCapture[] {
    const all: PacketCapture[] = [
      {
        id: 'pkt-001',
        timestamp: '12:44:02.105',
        srcIP: '198.51.100.12',
        dstIP: '10.244.0.15',
        srcPort: 53420,
        dstPort: 443,
        protocol: 'TCP',
        length: 512,
        action: 'allowed',
        hexPreview: '4500 0200 1c46 4000 4006 b1e6 c633 640c 0af4 000f',
      },
      {
        id: 'pkt-002',
        timestamp: '12:44:01.884',
        srcIP: '185.220.101.5',
        dstIP: '10.244.0.15',
        srcPort: 58921,
        dstPort: 22,
        protocol: 'TCP',
        length: 64,
        action: 'dropped',
        reason: 'StraitNetworkPolicy DROP',
        hexPreview: '4500 0040 a122 4000 3406 e411 b9dc 6505 0af4 000f',
      },
      {
        id: 'pkt-003',
        timestamp: '12:44:00.320',
        srcIP: '10.244.1.42',
        dstIP: '10.250.2.1',
        srcPort: 41200,
        dstPort: 51820,
        protocol: 'UDP',
        length: 1280,
        action: 'forwarded',
        hexPreview: '4500 0500 89a1 4000 4011 23f1 0af4 012a 0afa 0201',
      },
    ];

    return action ? all.filter(p => p.action === action) : all;
  }
}
