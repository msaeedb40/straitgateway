// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { describe, it, expect } from 'vitest';
import { calculateClusterHealth, calculateMapUtilization, getHealthStatusInfo } from './calculators';

describe('Dynamic Calculators Utility', () => {
  it('should compute 100% health when all resources are ready and active', () => {
    const health = calculateClusterHealth({
      readyNodes: 5,
      totalNodes: 5,
      activeGateways: 2,
      totalGateways: 2,
      activeTunnels: 4,
      totalTunnels: 4,
      droppedFlowsPerSec: 0,
      totalFlowsPerSec: 500,
    });
    expect(health).toBe(100);
    expect(getHealthStatusInfo(health).status).toBe('Good');
  });

  it('should dynamically degrade health when nodes or tunnels fail', () => {
    const health = calculateClusterHealth({
      readyNodes: 2,
      totalNodes: 4, // 50% node health
      activeGateways: 1,
      totalGateways: 2, // 50% gw health
      activeTunnels: 2,
      totalTunnels: 4, // 50% tunnel health
      droppedFlowsPerSec: 200,
      totalFlowsPerSec: 400,
    });
    expect(health).toBeLessThan(80);
    expect(getHealthStatusInfo(health).status).toBe('Critical');
  });

  it('should calculate accurate eBPF map utilization percentage', () => {
    expect(calculateMapUtilization(65536, 262144)).toBe(25);
    expect(calculateMapUtilization(0, 1000)).toBe(0);
    expect(calculateMapUtilization(1500, 1000)).toBe(100);
  });
});
