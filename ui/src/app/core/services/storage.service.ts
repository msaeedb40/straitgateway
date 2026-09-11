// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, PLATFORM_ID, inject } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';

@Injectable({
  providedIn: 'root',
})
export class StorageService {
  private platformId = inject(PLATFORM_ID);
  private isBrowser = isPlatformBrowser(this.platformId);

  getLocal<T>(key: string, defaultValue: T): T {
    if (!this.isBrowser) return defaultValue;
    try {
      const item = localStorage.getItem(key);
      return item ? (JSON.parse(item) as T) : defaultValue;
    } catch {
      return defaultValue;
    }
  }

  setLocal<T>(key: string, value: T): void {
    if (!this.isBrowser) return;
    try {
      localStorage.setItem(key, JSON.stringify(value));
    } catch (err) {
      console.warn('StorageService: failed to write to localStorage', err);
    }
  }

  removeLocal(key: string): void {
    if (!this.isBrowser) return;
    try {
      localStorage.removeItem(key);
    } catch {}
  }

  getSession<T>(key: string, defaultValue: T): T {
    if (!this.isBrowser) return defaultValue;
    try {
      const item = sessionStorage.getItem(key);
      return item ? (JSON.parse(item) as T) : defaultValue;
    } catch {
      return defaultValue;
    }
  }

  setSession<T>(key: string, value: T): void {
    if (!this.isBrowser) return;
    try {
      sessionStorage.setItem(key, JSON.stringify(value));
    } catch (err) {
      console.warn('StorageService: failed to write to sessionStorage', err);
    }
  }

  removeSession(key: string): void {
    if (!this.isBrowser) return;
    try {
      sessionStorage.removeItem(key);
    } catch {}
  }
}
