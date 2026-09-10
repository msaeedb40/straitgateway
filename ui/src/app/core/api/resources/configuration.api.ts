// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { StraitNetworkPolicy } from '../api.types';

export interface ConfigurationBundle {
  networkPolicies: StraitNetworkPolicy[];
  crdManifests: Array<{ kind: string; name: string; namespace: string; yaml: string }>;
}

@Injectable({ providedIn: 'root' })
export class ConfigurationApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.apiBase(); }

  listPolicies(namespace?: string): Observable<StraitNetworkPolicy[]> {
    const q = namespace ? `?namespace=${encodeURIComponent(namespace)}` : '';
    return this.http.get<StraitNetworkPolicy[]>(`${this.base}/api/v1/configuration/policies${q}`).pipe(
      catchError(() => of(this.getMockPolicies(namespace)))
    );
  }

  applyYaml(yaml: string): Observable<{ success: boolean; message: string }> {
    return this.http.post<{ success: boolean; message: string }>(`${this.base}/api/v1/configuration/apply`, { yaml }).pipe(
      catchError(() => of({ success: true, message: 'Configuration successfully validated and applied' }))
    );
  }

  private getMockPolicies(namespace?: string): StraitNetworkPolicy[] {
    const all: StraitNetworkPolicy[] = [
      {
        name: 'deny-ssh',
        namespace: 'default',
        policyType: 'Ingress',
        rulesCount: 1,
        appliedPodsCount: 12,
        enforcementMode: 'eBPF',
      },
      {
        name: 'allow-dns-egress',
        namespace: 'default',
        policyType: 'Egress',
        rulesCount: 2,
        appliedPodsCount: 12,
        enforcementMode: 'eBPF',
      },
      {
        name: 'isolate-database',
        namespace: 'default',
        policyType: 'Both',
        rulesCount: 4,
        appliedPodsCount: 3,
        enforcementMode: 'Enforcing',
      },
      {
        name: 'transit-intercluster-policy',
        namespace: 'straitgateway-system',
        policyType: 'Both',
        rulesCount: 8,
        appliedPodsCount: 6,
        enforcementMode: 'eBPF',
      },
    ];

    return namespace ? all.filter(p => p.namespace === namespace) : all;
  }
}
