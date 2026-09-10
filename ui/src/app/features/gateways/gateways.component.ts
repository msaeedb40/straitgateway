// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiClient, GatewaySummary } from '../../core/api/api-client';
import { NamespaceService } from '../../core/services/namespace.service';
import { catchError, of } from 'rxjs';

@Component({
  selector: 'app-gateways',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div>
      <h1 class="sg-page-title">Gateways</h1>
      <p class="sg-page-subtitle">Gateway API resources managed by straitgateway.io/skgateway</p>
    </div>
    <button class="sg-btn" (click)="showCreate.set(true)">+ New Gateway</button>
  </div>

  <!-- Create form -->
  @if (showCreate()) {
    <div class="sg-card" style="margin-bottom:24px;border-color:var(--sg-accent)">
      <div class="sg-card-header"><span class="sg-card-title">Create Gateway</span></div>
      <div class="sg-card-body">
        <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:16px;margin-bottom:16px">
          <div class="sg-form-group" style="margin-bottom:0">
            <label class="sg-form-label">Name</label>
            <input class="sg-input" style="width:100%" [(ngModel)]="form.name" placeholder="my-gateway">
          </div>
          <div class="sg-form-group" style="margin-bottom:0">
            <label class="sg-form-label">Namespace</label>
            <select class="sg-select" style="width:100%" [(ngModel)]="form.namespace">
              @for (ns of ns.namespaces(); track ns) { <option [value]="ns">{{ ns }}</option> }
            </select>
          </div>
          <div class="sg-form-group" style="margin-bottom:0">
            <label class="sg-form-label">Protocol</label>
            <select class="sg-select" style="width:100%" [(ngModel)]="form.protocol">
              <option>HTTP</option><option>HTTPS</option><option>TCP</option><option>UDP</option><option>TLS</option>
            </select>
          </div>
        </div>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px;margin-bottom:20px">
          <div class="sg-form-group" style="margin-bottom:0">
            <label class="sg-form-label">Port</label>
            <input class="sg-input" style="width:100%" type="number" [(ngModel)]="form.port" min="1" max="65535">
          </div>
        </div>
        <div style="display:flex;gap:10px">
          <button class="sg-btn" (click)="create()">Create</button>
          <button class="sg-btn sg-btn-secondary" (click)="showCreate.set(false);resetForm()">Cancel</button>
        </div>
      </div>
    </div>
  }

  <div class="sg-card">
    <div class="sg-card-header">
      <span class="sg-card-title">All Gateways</span>
      <span class="text-muted text-sm">{{ filtered().length }} gateways · namespace: {{ ns.active() || 'all' }}</span>
    </div>
    <div class="sg-card-body p-0">
      <table class="sg-table">
        <thead><tr>
          <th>Name</th><th>Namespace</th><th>Class</th><th>Addresses</th>
          <th>Listeners</th><th>Status</th><th style="width:120px">Actions</th>
        </tr></thead>
        <tbody>
          @if (loading()) {
            @for (i of [1,2,3]; track i) {
              <tr><td colspan="7"><div class="sg-skeleton" style="height:14px;width:70%;margin:2px 0"></div></td></tr>
            }
          } @else if (filtered().length === 0) {
            <tr><td colspan="7">
              <div class="sg-empty"><p>No gateways found</p></div>
            </td></tr>
          } @else {
            @for (gw of filtered(); track gw.name + gw.namespace) {
              <tr>
                <td><span class="mono text-sm">{{ gw.name }}</span></td>
                <td><span class="sg-badge pending">{{ gw.namespace }}</span></td>
                <td class="mono text-sm">{{ gw.gatewayClass }}</td>
                <td class="mono text-sm">{{ gw.addresses?.join(', ') || '—' }}</td>
                <td>{{ gw.listeners }}</td>
                <td><span class="sg-badge" [class]="gw.ready ? 'active' : 'warn'">{{ gw.ready ? 'Ready' : 'Pending' }}</span></td>
                <td>
                  <div style="display:flex;gap:6px">
                    <button class="sg-btn sg-btn-secondary sg-btn-sm"
                      (click)="viewYaml(gw)">YAML</button>
                    <button class="sg-btn sg-btn-secondary sg-btn-sm" style="color:var(--sg-danger)"
                      (click)="confirmDelete(gw)">Delete</button>
                  </div>
                </td>
              </tr>
            }
          }
        </tbody>
      </table>
    </div>
  </div>

  <!-- YAML viewer -->
  @if (selectedYaml()) {
    <div class="sg-card" style="margin-top:24px">
      <div class="sg-card-header">
        <span class="sg-card-title">{{ selectedGw()?.name }} — YAML</span>
        <button class="sg-btn sg-btn-secondary sg-btn-sm" (click)="selectedYaml.set('')">Close</button>
      </div>
      <div class="sg-card-body">
        <pre class="sg-yaml-code"><code>{{ selectedYaml() }}</code></pre>
      </div>
    </div>
  }

  <!-- Delete confirm -->
  @if (deleteTarget()) {
    <div class="sg-dialog-overlay" (click)="deleteTarget.set(null)">
      <div class="sg-dialog" (click)="$event.stopPropagation()">
        <div style="font-size:18px;font-weight:700;color:var(--sg-text)">Delete Gateway</div>
        <div style="font-size:13px;color:var(--sg-text-2);line-height:1.5">
          Delete <strong style="color:var(--sg-text)">{{ deleteTarget()?.name }}</strong> in namespace <strong style="color:var(--sg-text)">{{ deleteTarget()?.namespace }}</strong>? This cannot be undone.
        </div>
        <div class="sg-dialog-actions">
          <button class="sg-btn sg-btn-secondary" (click)="deleteTarget.set(null)">Cancel</button>
          <button class="sg-btn" style="background:var(--sg-danger);color:#fff" (click)="doDelete()">Delete</button>
        </div>
      </div>
    </div>
  }
</div>
`,
})
export class GatewaysComponent implements OnInit {
  private api = inject(ApiClient);
  protected ns = inject(NamespaceService);

  gateways = signal<GatewaySummary[]>([]);
  loading  = signal(true);
  showCreate = signal(false);
  selectedGw   = signal<GatewaySummary | null>(null);
  selectedYaml = signal('');
  deleteTarget = signal<GatewaySummary | null>(null);

  form = { name: '', namespace: 'default', protocol: 'HTTP', port: 80 };

  filtered() {
    const ns = this.ns.active();
    return ns ? this.gateways().filter(g => g.namespace === ns) : this.gateways();
  }

  ngOnInit() { this.load(); }

  load() {
    this.loading.set(true);
    this.api.getGateways().pipe(catchError(() => of([]))).subscribe(data => {
      this.gateways.set(data);
      this.loading.set(false);
    });
  }

  create() {
    const yaml = this.buildGatewayYaml(this.form);
    console.log('[straitgateway] Create gateway:\n', yaml);
    this.showCreate.set(false);
    this.resetForm();
  }

  viewYaml(gw: GatewaySummary) {
    this.selectedGw.set(gw);
    this.selectedYaml.set(this.buildExistingGatewayYaml(gw));
  }

  confirmDelete(gw: GatewaySummary) { this.deleteTarget.set(gw); }
  doDelete() {
    console.log('[straitgateway] Delete gateway:', this.deleteTarget());
    this.gateways.update(list => list.filter(g => g !== this.deleteTarget()));
    this.deleteTarget.set(null);
  }

  resetForm() { this.form = { name: '', namespace: 'default', protocol: 'HTTP', port: 80 }; }

  private buildGatewayYaml(f: typeof this.form): string {
    return `apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: ${f.name}
  namespace: ${f.namespace}
spec:
  gatewayClassName: skgateway
  listeners:
    - name: ${f.protocol.toLowerCase()}
      port: ${f.port}
      protocol: ${f.protocol}`;
  }

  private buildExistingGatewayYaml(gw: GatewaySummary): string {
    return `apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: ${gw.name}
  namespace: ${gw.namespace}
spec:
  gatewayClassName: ${gw.gatewayClass}
  listeners:
    - name: default
      port: 80
      protocol: HTTP
status:
  addresses:
${(gw.addresses ?? []).map(a => `    - value: ${a}`).join('\n') || '    []'}`;
  }
}
