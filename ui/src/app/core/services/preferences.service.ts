import { Injectable, PLATFORM_ID, inject } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';

@Injectable({ providedIn: 'root' })
export class PreferencesService {
  private readonly platformId = inject(PLATFORM_ID);

  get<T>(key: string, defaultValue: T): T {
    if (!isPlatformBrowser(this.platformId)) return defaultValue;
    try {
      const raw = localStorage.getItem(`sg:pref:${key}`);
      return raw !== null ? (JSON.parse(raw) as T) : defaultValue;
    } catch {
      return defaultValue;
    }
  }

  set<T>(key: string, value: T): void {
    if (!isPlatformBrowser(this.platformId)) return;
    try {
      localStorage.setItem(`sg:pref:${key}`, JSON.stringify(value));
    } catch {
      // Storage quota exceeded or private mode — silently ignore
    }
  }

  remove(key: string): void {
    if (!isPlatformBrowser(this.platformId)) return;
    localStorage.removeItem(`sg:pref:${key}`);
  }
}
