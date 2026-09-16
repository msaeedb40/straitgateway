import { Component, inject, signal, DestroyRef, PLATFORM_ID } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { isPlatformBrowser, SlicePipe } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { PacketApiService } from '../../core/api/resources/packet.api';
import { BreadcrumbsComponent } from '../../layout/breadcrumbs/breadcrumbs.component';
import { StatusBadgeComponent } from '../../shared/components/status-badge/status-badge.component';
import { ErrorStateComponent } from '../../shared/components/error-state/error-state.component';
import { CaptureSession, CapturedPacket, CaptureProtocol } from '../../core/models/packet.model';
import { ApiError } from '../../core/api/api-error';

const PROTOCOLS: CaptureProtocol[] = ['TCP', 'UDP', 'ICMP', 'ARP', 'ANY'];

@Component({
  selector: 'sg-packets',
  imports: [ReactiveFormsModule, SlicePipe, BreadcrumbsComponent, StatusBadgeComponent, ErrorStateComponent],
  template: `
    <div class="sg-page">
      <sg-breadcrumbs />
      <header class="sg-page-header">
        <h1 class="sg-page-title">Packet Capture</h1>
      </header>

      <!-- Capture configuration form -->
      <section class="sg-card" aria-labelledby="capture-form-heading">
        <h2 id="capture-form-heading" class="sg-section-title">Capture Configuration</h2>
        <form [formGroup]="form" (ngSubmit)="startCapture()" class="sg-capture-form" novalidate>
          <div class="sg-form-grid">
            <div class="sg-form-field">
              <label for="capture-node" class="sg-form-label">Node <span aria-hidden="true">*</span></label>
              <input id="capture-node" class="sg-input" formControlName="nodeName" placeholder="Enter node name" [attr.aria-required]="true" />
              @if (form.get('nodeName')?.invalid && form.get('nodeName')?.touched) {
                <span class="sg-form-error" role="alert">Node name is required</span>
              }
            </div>
            <div class="sg-form-field">
              <label for="capture-ns" class="sg-form-label">Namespace</label>
              <input id="capture-ns" class="sg-input" formControlName="namespace" placeholder="e.g. kube-system" />
            </div>
            <div class="sg-form-field">
              <label for="capture-pod" class="sg-form-label">Pod (optional)</label>
              <input id="capture-pod" class="sg-input" formControlName="podName" placeholder="Enter pod name" />
            </div>
            <div class="sg-form-field">
              <label for="capture-iface" class="sg-form-label">Interface (optional)</label>
              <input id="capture-iface" class="sg-input" formControlName="interfaceName" placeholder="e.g. eth0" />
            </div>
            <div class="sg-form-field">
              <label for="capture-proto" class="sg-form-label">Protocol</label>
              <select id="capture-proto" class="sg-input sg-select" formControlName="protocol">
                @for (p of protocols; track p) {
                  <option [value]="p">{{ p }}</option>
                }
              </select>
            </div>
            <div class="sg-form-field">
              <label for="capture-src" class="sg-form-label">Source filter</label>
              <input id="capture-src" class="sg-input" formControlName="sourceFilter" placeholder="e.g. 10.0.0.1" />
            </div>
            <div class="sg-form-field">
              <label for="capture-dst" class="sg-form-label">Destination filter</label>
              <input id="capture-dst" class="sg-input" formControlName="destinationFilter" placeholder="e.g. 10.0.0.2" />
            </div>
            <div class="sg-form-field">
              <label for="capture-port" class="sg-form-label">Port (optional)</label>
              <input id="capture-port" class="sg-input" type="number" formControlName="port" placeholder="e.g. 443" min="1" max="65535" />
            </div>
            <div class="sg-form-field">
              <label for="capture-duration" class="sg-form-label">Duration (seconds)</label>
              <input id="capture-duration" class="sg-input" type="number" formControlName="durationSeconds" min="1" max="3600" />
            </div>
            <div class="sg-form-field">
              <label for="capture-limit" class="sg-form-label">Packet limit</label>
              <input id="capture-limit" class="sg-input" type="number" formControlName="packetLimit" min="1" max="100000" />
            </div>
          </div>

          <div class="sg-capture-actions">
            @if (!session() || session()?.state === 'Stopped' || session()?.state === 'Completed' || session()?.state === 'Failed') {
              <button class="sg-btn sg-btn-primary" type="submit" [disabled]="form.invalid || starting()">
                {{ starting() ? 'Starting…' : 'Start Capture' }}
              </button>
            } @else {
              <button class="sg-btn sg-btn-danger" type="button" (click)="stopCapture()">Stop Capture</button>
            }
          </div>
          @if (formError()) {
            <sg-error-state [error]="formError()" />
          }
        </form>
      </section>

      <!-- Live packet stream -->
      @if (session()) {
        <section class="sg-card" aria-labelledby="capture-stream-heading" aria-live="polite">
          <div class="sg-capture-session-header">
            <h2 id="capture-stream-heading" class="sg-section-title">
              Packet Stream
            </h2>
            <sg-status-badge
              [variant]="session()!.state === 'Running' ? 'healthy' : session()!.state === 'Failed' ? 'failed' : 'unknown'"
            />
            <span class="sg-capture-count" aria-live="polite" aria-atomic="true">
              {{ packets().length }} packets
            </span>
          </div>

          @if (session()!.error) {
            <p class="sg-capture-error" role="alert">{{ session()!.error }}</p>
          }

          <div class="sg-table-wrapper" role="region" aria-label="Captured packets">
            <table class="sg-table" aria-rowcount="{{ packets().length }}">
              <thead>
                <tr>
                  <th scope="col">#</th>
                  <th scope="col">Time</th>
                  <th scope="col">Source</th>
                  <th scope="col">Destination</th>
                  <th scope="col">Protocol</th>
                  <th scope="col">Length</th>
                  <th scope="col">Direction</th>
                </tr>
              </thead>
              <tbody>
                @for (pkt of packets(); track pkt.index) {
                  <tr [attr.aria-rowindex]="pkt.index">
                    <td class="sg-mono-value">{{ pkt.index }}</td>
                    <td class="sg-mono-value">{{ pkt.timestamp | slice:11:23 }}</td>
                    <td class="sg-mono-value">{{ pkt.sourceIP }}{{ pkt.sourcePort !== null ? ':' + pkt.sourcePort : '' }}</td>
                    <td class="sg-mono-value">{{ pkt.destinationIP }}{{ pkt.destinationPort !== null ? ':' + pkt.destinationPort : '' }}</td>
                    <td>{{ pkt.protocol }}</td>
                    <td class="sg-mono-value">{{ pkt.length }}</td>
                    <td>{{ pkt.direction }}</td>
                  </tr>
                }
              </tbody>
            </table>
          </div>
        </section>
      }
    </div>
  `,
  styles: [`
    .sg-section-title { font-size:.8125rem;font-weight:600;color:var(--sg-text-secondary);text-transform:uppercase;letter-spacing:.05em;margin-bottom:1rem; }
    .sg-form-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(220px,1fr));gap:12px;margin-bottom:1rem; }
    .sg-form-field { display:flex;flex-direction:column;gap:4px; }
    .sg-form-label { font-size:.75rem;font-weight:500;color:var(--sg-text-secondary); }
    .sg-form-error { font-size:.75rem;color:var(--sg-failed); }
    .sg-capture-actions { display:flex;gap:8px;margin-top:1rem; }
    .sg-capture-session-header { display:flex;align-items:center;gap:12px;margin-bottom:1rem; }
    .sg-capture-count { font-size:.75rem;color:var(--sg-text-muted);font-family:var(--sg-font-mono); }
    .sg-capture-error { color:var(--sg-failed);font-size:.8125rem;margin-bottom:.5rem; }
  `],
})
export class PacketsComponent {
  private readonly packetApi  = inject(PacketApiService);
  private readonly platformId = inject(PLATFORM_ID);
  private readonly destroyRef = inject(DestroyRef);
  private readonly fb         = inject(FormBuilder);

  readonly protocols = PROTOCOLS;
  readonly session   = signal<CaptureSession | null>(null);
  readonly packets   = signal<CapturedPacket[]>([]);
  readonly starting  = signal(false);
  readonly formError = signal<ApiError | null>(null);

  private eventSource: EventSource | null = null;

  readonly form = this.fb.nonNullable.group({
    nodeName:          ['', Validators.required],
    namespace:         [''],
    podName:           [''],
    interfaceName:     [''],
    protocol:          ['ANY' as CaptureProtocol],
    sourceFilter:      [''],
    destinationFilter: [''],
    port:              [null as number | null],
    durationSeconds:   [30],
    packetLimit:       [1000],
  });

  startCapture(): void {
    if (this.form.invalid) return;
    this.starting.set(true);
    this.formError.set(null);
    this.packets.set([]);

    const raw = this.form.getRawValue();
    const spec = {
      nodeName:          raw.nodeName,
      namespace:         raw.namespace || undefined,
      podName:           raw.podName   || undefined,
      interfaceName:     raw.interfaceName || undefined,
      protocol:          raw.protocol,
      sourceFilter:      raw.sourceFilter  || undefined,
      destinationFilter: raw.destinationFilter || undefined,
      port:              raw.port ?? undefined,
      durationSeconds:   raw.durationSeconds ?? undefined,
      packetLimit:       raw.packetLimit ?? undefined,
    };

    this.packetApi.createSession(spec)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (s) => {
          this.session.set(s);
          this.starting.set(false);
          this.openStream(s.id);
        },
        error: (err) => { this.formError.set(err); this.starting.set(false); },
      });
  }

  stopCapture(): void {
    const s = this.session();
    if (!s) return;
    this.packetApi.stopSession(s.id)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({ next: (updated) => { this.session.set(updated); this.closeStream(); } });
  }

  private openStream(sessionId: string): void {
    if (!isPlatformBrowser(this.platformId)) return;
    this.closeStream();
    const url = this.packetApi.streamUrl(sessionId);
    this.eventSource = new EventSource(url);
    this.eventSource.onmessage = (e) => {
      const pkt: CapturedPacket = JSON.parse(e.data);
      this.packets.update((list) => [...list, pkt]);
    };
    this.eventSource.onerror = () => this.closeStream();
  }

  private closeStream(): void {
    this.eventSource?.close();
    this.eventSource = null;
  }
}
