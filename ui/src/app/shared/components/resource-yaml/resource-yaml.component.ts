// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';

/** Renders a Kubernetes resource as YAML in a code block. For CRUD operations. */
@Component({
  selector: 'sg-resource-yaml',
  standalone: true,
  imports: [CommonModule],
  template: `
<div class="sg-yaml-viewer">
  <div class="sg-yaml-header">
    <span class="sg-yaml-title">{{ title() }}</span>
    <button class="sg-btn sg-btn-secondary" style="padding:2px 8px;font-size:11px" (click)="copyYaml()">Copy</button>
  </div>
  <pre class="sg-yaml-code"><code>{{ yaml() }}</code></pre>
</div>
`,
  styles: [`
    .sg-yaml-viewer {
      border: 1px solid var(--sg-border); border-radius: var(--sg-radius-md);
      overflow: hidden; background: var(--sg-surface-1);
    }
    .sg-yaml-header {
      display: flex; justify-content: space-between; align-items: center;
      padding: 8px 12px; background: var(--sg-surface-2);
      border-bottom: 1px solid var(--sg-border); font-size: 12px;
    }
    .sg-yaml-title { color: var(--sg-text-2); font-weight: 500; }
    .sg-yaml-code {
      margin: 0; padding: 16px; font-size: 12px; line-height: 1.6;
      color: var(--sg-text-1); overflow-x: auto; font-family: var(--sg-font-mono);
    }
  `],
})
export class ResourceYamlComponent {
  title = input<string>('Resource YAML');
  yaml  = input.required<string>();

  async copyYaml() {
    try { await navigator.clipboard.writeText(this.yaml()); } catch {}
  }
}
