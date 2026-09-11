// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

/**
 * Dynamically computes cluster network health percentage based on live resource states.
 */
export function calculateClusterHealth(params: {
  readyNodes: number;
  totalNodes: number;
  activeGateways: number;
  totalGateways: number;
  activeTunnels: number;
  totalTunnels: number;
  droppedFlowsPerSec?: number;
  totalFlowsPerSec?: number;
}): number {
  let score = 0;
  let weights = 0;

  // Node health weight (40%)
  if (params.totalNodes > 0) {
    const nodeHealth = (params.readyNodes / params.totalNodes) * 100;
    score += nodeHealth * 0.4;
    weights += 0.4;
  }

  // Gateway health weight (30%)
  if (params.totalGateways > 0) {
    const gwHealth = (params.activeGateways / params.totalGateways) * 100;
    score += gwHealth * 0.3;
    weights += 0.3;
  }

  // Tunnel health weight (20%)
  if (params.totalTunnels > 0) {
    const tunnelHealth = (params.activeTunnels / params.totalTunnels) * 100;
    score += tunnelHealth * 0.2;
    weights += 0.2;
  }

  // Flow drop penalty (10%)
  if (params.totalFlowsPerSec && params.totalFlowsPerSec > 0) {
    const dropRate = (params.droppedFlowsPerSec || 0) / params.totalFlowsPerSec;
    const dropHealth = Math.max(0, 100 - dropRate * 200);
    score += dropHealth * 0.1;
    weights += 0.1;
  }

  if (weights === 0) return 100.0;
  return Math.min(100, Math.max(0, Math.round((score / weights) * 10) / 10));
}

/**
 * Dynamically computes eBPF map capacity percentage.
 */
export function calculateMapUtilization(current: number, max: number): number {
  if (max <= 0) return 0;
  return Math.min(100, Math.round((current / max) * 1000) / 10);
}

/**
 * Returns health category label and CSS color class dynamically.
 */
export function getHealthStatusInfo(healthPercentage: number): {
  status: 'Good' | 'Degraded' | 'Critical';
  colorClass: string;
  badgeClass: string;
} {
  if (healthPercentage >= 95) {
    return { status: 'Good', colorClass: 'text-success', badgeClass: 'active' };
  } else if (healthPercentage >= 80) {
    return { status: 'Degraded', colorClass: 'text-warn', badgeClass: 'warn' };
  } else {
    return { status: 'Critical', colorClass: 'text-danger', badgeClass: 'inactive' };
  }
}
