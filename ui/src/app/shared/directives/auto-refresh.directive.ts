// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Directive, OnInit, OnDestroy, input, output } from '@angular/core';

/** Emits a refresh event at a configurable interval (ms). Default: 15000. */
@Directive({ selector: '[sgAutoRefresh]', standalone: true })
export class AutoRefreshDirective implements OnInit, OnDestroy {
  sgAutoRefresh = input<number>(15000);
  refresh = output<void>();
  private intervalId: ReturnType<typeof setInterval> | null = null;

  ngOnInit() {
    const ms = this.sgAutoRefresh();
    if (ms > 0) {
      this.intervalId = setInterval(() => this.refresh.emit(), ms);
    }
  }

  ngOnDestroy() {
    if (this.intervalId) clearInterval(this.intervalId);
  }
}
