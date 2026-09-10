// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, signal, effect } from '@angular/core';

export interface UserPreferences {
  theme: 'dark' | 'light';
  refreshInterval: number; // in seconds
  compactTables: boolean;
  defaultNamespace: string;
  enableSoundAlerts: boolean;
  enableMockFallback: boolean;
}

const STORAGE_KEY = 'straitgateway_prefs_v1';

const DEFAULT_PREFS: UserPreferences = {
  theme: 'dark',
  refreshInterval: 15,
  compactTables: false,
  defaultNamespace: 'default',
  enableSoundAlerts: false,
  enableMockFallback: true,
};

@Injectable({ providedIn: 'root' })
export class PreferencesService {
  private readonly _prefs = signal<UserPreferences>(this.loadFromStorage());
  readonly prefs = this._prefs.asReadonly();

  constructor() {
    effect(() => {
      const p = this._prefs();
      if (typeof window !== 'undefined' && window.localStorage) {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(p));
      }
    });
  }

  update(partial: Partial<UserPreferences>) {
    this._prefs.update((cur) => ({ ...cur, ...partial }));
  }

  private loadFromStorage(): UserPreferences {
    if (typeof window === 'undefined' || !window.localStorage) {
      return DEFAULT_PREFS;
    }
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      return raw ? { ...DEFAULT_PREFS, ...JSON.parse(raw) } : DEFAULT_PREFS;
    } catch {
      return DEFAULT_PREFS;
    }
  }
}
