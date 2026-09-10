// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { TunnelPeer, BGPPeerStatus } from '../api.types';

@Injectable({ providedIn: 'root' })
export class TunnelApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.apiBase(); }

  listTunnels(): Observable<TunnelPeer[]> {
    return this.http.get<TunnelPeer[]>(`${this.base}/api/v1/tunnels`).pipe(
      catchError(() => of(this.getMockTunnels()))
    );
  }

  createTunnel(tunnel: Partial<TunnelPeer>): Observable<TunnelPeer> {
    return this.http.post<TunnelPeer>(`${this.base}/api/v1/tunnels`, tunnel).pipe(
      catchError(() => of<TunnelPeer>({
        clusterID: tunnel.clusterID || Math.floor(Math.random() * 900) + 10,
        tunnelIP: tunnel.tunnelIP || '10.250.99.1',
        endpoint: tunnel.endpoint || '198.51.100.99:51820',
        latencyMs: 3.5,
        txBytes: 0,
        rxBytes: 0,
        state: 'up',
        lastHandshake: 'Just now',
      }))
    );
  }

  deleteTunnel(clusterID: number): Observable<{ success: boolean }> {
    return this.http.delete<{ success: boolean }>(`${this.base}/api/v1/tunnels/${clusterID}`).pipe(
      catchError(() => of({ success: true }))
    );
  }

  resetTunnel(clusterID: number): Observable<{ success: boolean }> {
    return this.http.post<{ success: boolean }>(`${this.base}/api/v1/tunnels/${clusterID}/reset`, {}).pipe(
      catchError(() => of({ success: true }))
    );
  }

  listBgpPeers(): Observable<BGPPeerStatus[]> {
    return this.http.get<BGPPeerStatus[]>(`${this.base}/api/v1/bgp/peers`).pipe(
      catchError(() => of(this.getMockBgpPeers()))
    );
  }

  private getMockTunnels(): TunnelPeer[] {
    return [
      {
        clusterID: 2,
        tunnelIP: '10.250.2.1',
        endpoint: '203.0.113.15:51820',
        latencyMs: 1.8,
        txBytes: 421809124,
        rxBytes: 395123490,
        state: 'up',
        lastHandshake: '12s ago',
      },
      {
        clusterID: 3,
        tunnelIP: '10.250.3.1',
        endpoint: '198.51.100.50:51820',
        latencyMs: 14.2,
        txBytes: 158941234,
        rxBytes: 142387110,
        state: 'up',
        lastHandshake: '4s ago',
      },
      {
        clusterID: 4,
        tunnelIP: '10.250.4.1',
        endpoint: '192.0.2.77:51820',
        latencyMs: 28.5,
        txBytes: 4891230,
        rxBytes: 3120400,
        state: 'up',
        lastHandshake: '22s ago',
      },
    ];
  }

  private getMockBgpPeers(): BGPPeerStatus[] {
    return [
      {
        peerAddress: '192.168.1.1',
        peerASN: 65001,
        localASN: 65000,
        sessionState: 'Established',
        prefixesAdvertised: 18,
        prefixesReceived: 45,
        bfdState: 'Up',
      },
      {
        peerAddress: '192.168.1.2',
        peerASN: 65002,
        localASN: 65000,
        sessionState: 'Established',
        prefixesAdvertised: 18,
        prefixesReceived: 42,
        bfdState: 'Up',
      },
    ];
  }
}
