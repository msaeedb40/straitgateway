import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { PaginatedResponse } from '../api.types';
import { CaptureSession, CaptureSpec, CapturedPacket } from '../../models/packet.model';

@Injectable({ providedIn: 'root' })
export class PacketApiService {
  private readonly client = inject(ApiClient);

  createSession(spec: CaptureSpec): Observable<CaptureSession> {
    return this.client.post<CaptureSession>('/v1/packet-captures', { spec });
  }

  getSession(id: string): Observable<CaptureSession> {
    return this.client.get<CaptureSession>(`/v1/packet-captures/${id}`);
  }

  listSessions(): Observable<PaginatedResponse<CaptureSession>> {
    return this.client.get<PaginatedResponse<CaptureSession>>('/v1/packet-captures');
  }

  stopSession(id: string): Observable<CaptureSession> {
    return this.client.post<CaptureSession>(`/v1/packet-captures/${id}/stop`, {});
  }

  deleteSession(id: string): Observable<void> {
    return this.client.delete<void>(`/v1/packet-captures/${id}`);
  }

  getPackets(sessionId: string, page: number, pageSize: number): Observable<PaginatedResponse<CapturedPacket>> {
    return this.client.get<PaginatedResponse<CapturedPacket>>(
      `/v1/packet-captures/${sessionId}/packets`,
      { params: { page, pageSize } }
    );
  }

  /** Returns SSE URL for live packet streaming — caller uses EventSource */
  streamUrl(sessionId: string): string {
    return `${this.client.controllerBase}/v1/packet-captures/${sessionId}/stream`;
  }
}
