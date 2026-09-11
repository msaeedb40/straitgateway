// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-footer',
  standalone: true,
  imports: [CommonModule],
  template: `
    <footer class="sg-footer">
      <div class="sg-footer-left">
        <span class="sg-footer-brand">StraitGateway</span>
        <span class="sg-footer-version">v1.26.7-ebpf</span>
        <span class="sg-footer-dot">·</span>
        <span class="sg-footer-kernel">Linux 6.6+ LTS Verified</span>
      </div>
      <div class="sg-footer-right">
        <a href="https://straitgateway.io/docs" target="_blank" rel="noopener noreferrer" class="sg-footer-link">Docs</a>
        <a href="https://github.com/straitgateway" target="_blank" rel="noopener noreferrer" class="sg-footer-link">GitHub</a>
        <span class="sg-footer-status">
          <span class="sg-status-dot"></span>
          Dataplane Operational
        </span>
      </div>
    </footer>
  `,
  styles: [`
    .sg-footer {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 8px 18px;
      font-size: 11px;
      color: var(--sg-text-3);
      border-top: 1px solid var(--sg-border);
      background: rgba(13, 16, 24, 0.85);
      backdrop-filter: blur(10px);
    }
    .sg-footer-left, .sg-footer-right {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .sg-footer-brand {
      font-weight: 600;
      color: var(--sg-text-2);
    }
    .sg-footer-version {
      font-family: monospace;
      color: var(--sg-accent-light);
    }
    .sg-footer-link {
      color: var(--sg-text-3);
      transition: color 150ms ease;
    }
    .sg-footer-link:hover {
      color: var(--sg-text);
    }
    .sg-footer-status {
      display: flex;
      align-items: center;
      gap: 5px;
      color: var(--sg-success);
      margin-left: 8px;
    }
    .sg-status-dot {
      width: 5px;
      height: 5px;
      border-radius: 50%;
      background: var(--sg-success);
      box-shadow: 0 0 5px var(--sg-success);
    }
  `],
})
export class FooterComponent {}
