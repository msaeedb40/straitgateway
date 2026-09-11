// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

/**
 * Formats a raw byte count into human-readable unit (B, KB, MB, GB, TB).
 */
export function formatBytes(bytes: number, decimals: number = 1): string {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
}

/**
 * Formats rate (e.g. bytes/sec or packets/sec).
 */
export function formatRate(val: number, unit: string = 'pps'): string {
  if (val >= 1_000_000) return `${(val / 1_000_000).toFixed(1)} M${unit}`;
  if (val >= 1_000) return `${(val / 1_000).toFixed(1)} K${unit}`;
  return `${Math.round(val)} ${unit}`;
}

/**
 * Formats duration in milliseconds or nanoseconds.
 */
export function formatDurationMs(ms: number): string {
  if (ms < 1) return `${(ms * 1000).toFixed(0)} µs`;
  if (ms < 1000) return `${ms.toFixed(1)} ms`;
  return `${(ms / 1000).toFixed(2)} s`;
}

/**
 * Formats relative time elapsed since a timestamp.
 */
export function formatTimeAgo(timestamp: string | number): string {
  const date = typeof timestamp === 'number' ? new Date(timestamp) : new Date(timestamp);
  const now = new Date();
  const diffSec = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (isNaN(diffSec) || diffSec < 0) return 'just now';
  if (diffSec < 60) return `${diffSec}s ago`;
  const diffMin = Math.floor(diffSec / 60);
  if (diffMin < 60) return `${diffMin}m ago`;
  const diffHours = Math.floor(diffMin / 60);
  if (diffHours < 24) return `${diffHours}h ago`;
  return `${Math.floor(diffHours / 24)}d ago`;
}
