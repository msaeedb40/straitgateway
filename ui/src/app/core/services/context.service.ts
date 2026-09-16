import { effect, inject, Injectable, signal } from '@angular/core';
import { PreferencesService } from './preferences.service';

@Injectable({ providedIn: 'root' })
export class ContextService {
  private readonly preferences = inject(PreferencesService);

  readonly selectedCluster = signal<string | null>(
    this.preferences.get<string | null>('selectedCluster', null)
  );

  readonly selectedNamespace = signal<string>(
    this.preferences.get<string>('selectedNamespace', '')
  );

  constructor() {
    // Persist context changes without hardcoding a default namespace
    effect(() => {
      this.preferences.set('selectedCluster', this.selectedCluster());
    });
    effect(() => {
      this.preferences.set('selectedNamespace', this.selectedNamespace());
    });
  }

  setCluster(cluster: string | null): void {
    this.selectedCluster.set(cluster);
    // Changing cluster clears namespace — new namespace list will be loaded by consumers
    this.selectedNamespace.set('');
  }

  setNamespace(ns: string): void {
    this.selectedNamespace.set(ns);
  }
}
