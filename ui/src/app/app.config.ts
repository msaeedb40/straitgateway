// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import {
  ApplicationConfig,
  provideZonelessChangeDetection,
  APP_INITIALIZER,
  inject,
} from '@angular/core';
import { provideRouter, withViewTransitions, withComponentInputBinding } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { routes } from './app.routes';
import { errorInterceptor } from './core/interceptors/error.interceptor';
import { RuntimeConfigService } from './core/config/runtime-config';

function initRuntimeConfig() {
  const config = inject(RuntimeConfigService);
  return () => config.load();
}

export const appConfig: ApplicationConfig = {
  providers: [
    provideZonelessChangeDetection(),
    provideRouter(routes, withViewTransitions(), withComponentInputBinding()),
    provideHttpClient(withInterceptors([errorInterceptor])),
    {
      provide: APP_INITIALIZER,
      useFactory: initRuntimeConfig,
      multi: true,
    },
  ],
};
