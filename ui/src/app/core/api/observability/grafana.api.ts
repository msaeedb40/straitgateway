// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { RuntimeConfigService } from '../../config/runtime-config';

export interface GrafanaDashboardRef {
  id: string;
  title: string;
  url: string;
  embedUrl: string;
}

@Injectable({ providedIn: 'root' })
export class GrafanaApi {
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.grafanaBase(); }

  getDashboards(): GrafanaDashboardRef[] {
    return [
      {
        id: 'sg-cluster-overview',
        title: 'StraitGateway Cluster Networking',
        url: `${this.base}/d/sg-cluster-overview/straitgateway-cluster-networking`,
        embedUrl: `${this.base}/d-solo/sg-cluster-overview?orgId=1&panelId=1`,
      },
      {
        id: 'sg-ebpf-dataplane',
        title: 'eBPF Dataplane & Map Limits',
        url: `${this.base}/d/sg-ebpf/straitgateway-ebpf-dataplane`,
        embedUrl: `${this.base}/d-solo/sg-ebpf?orgId=1&panelId=2`,
      },
      {
        id: 'sg-transit-tunnels',
        title: 'Transit Gateway WireGuard Latency',
        url: `${this.base}/d/sg-transit/straitgateway-transit-tunnels`,
        embedUrl: `${this.base}/d-solo/sg-transit?orgId=1&panelId=3`,
      },
    ];
  }
}
