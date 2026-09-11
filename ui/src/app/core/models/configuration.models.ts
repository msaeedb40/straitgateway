// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

export interface SubsystemStatus {
  name: string;
  category: 'dataplane' | 'cni' | 'routing' | 'gateway' | 'observability';
  status: 'Active' | 'Degraded' | 'Inactive';
  details?: string;
  lastUpdated?: string;
}

export interface UiPreferences {
  theme: 'dark' | 'light' | 'system';
  refreshIntervalMs: number;
  activeNamespace: string;
  showPacketHex: boolean;
  topologyLayout: 'force' | 'radial' | 'hierarchical';
  density: 'compact' | 'comfortable';
}

export interface SystemSettings {
  general: {
    clusterName: string;
    environment: string;
    defaultNamespace: string;
    logLevel: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';
  };
  gateway: {
    gatewayClassName: string;
    controllerName: string;
    enableBgpSync: boolean;
    defaultHttpPort: number;
    defaultHttpsPort: number;
  };
  network: {
    clusterPodCIDR: string;
    serviceCIDR: string;
    mtuDiscovery: boolean;
    tunnelMode: 'wireguard' | 'vxlan' | 'geneve' | 'direct';
    enableIPv6: boolean;
  };
  ebpf: {
    bpfFsPath: string;
    ringBufferSizeMb: number;
    conntrackMaxEntries: number;
    enableMaglevHash: boolean;
    enableDsr: boolean;
  };
  cni: {
    cniBinDir: string;
    cniConfDir: string;
    interfacePrefix: string;
    ipamBackend: 'strait-ipam' | 'host-local';
  };
  observability: {
    prometheusEndpoint: string;
    grafanaEndpoint: string;
    jaegerEndpoint: string;
    metricsScrapeIntervalSeconds: number;
  };
  security: {
    enforceNetworkPolicy: boolean;
    defaultPolicyAction: 'allow' | 'deny';
    wireguardKeyRotationDays: number;
    mTLSStrict: boolean;
  };
}
