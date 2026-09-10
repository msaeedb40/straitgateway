// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Pipe, PipeTransform } from '@angular/core';

/** Converts an ISO timestamp or Date to a relative time string (e.g. "3m ago"). */
@Pipe({ name: 'relativeTime', standalone: true })
export class RelativeTimePipe implements PipeTransform {
  transform(value: string | Date | null): string {
    if (!value) return '—';
    const now = Date.now();
    const then = new Date(value).getTime();
    const diff = Math.max(0, now - then);
    const sec = Math.floor(diff / 1000);
    if (sec < 60) return `${sec}s ago`;
    const min = Math.floor(sec / 60);
    if (min < 60) return `${min}m ago`;
    const hr = Math.floor(min / 60);
    if (hr < 24) return `${hr}h ago`;
    const days = Math.floor(hr / 24);
    return `${days}d ago`;
  }
}
