import {
  ApplicationConfig,
  provideBrowserGlobalErrorListeners,
  provideZonelessChangeDetection,
  APP_INITIALIZER,
} from '@angular/core';
import { provideRouter, withComponentInputBinding, withViewTransitions } from '@angular/router';
import { provideClientHydration, withEventReplay } from '@angular/platform-browser';
import {
  provideHttpClient,
  withInterceptors,
  withFetch,
  withXsrfConfiguration,
} from '@angular/common/http';

import { routes } from './app.routes';
import { RuntimeConfigService } from './core/config/runtime-config.service';
import { authInterceptor } from './core/interceptors/auth.interceptor';
import { errorInterceptor } from './core/interceptors/error.interceptor';
import { namespaceInterceptor } from './core/interceptors/namespace.interceptor';

function initializeApp(runtimeConfig: RuntimeConfigService): () => Promise<void> {
  return () => runtimeConfig.load();
}

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideZonelessChangeDetection(),
    provideRouter(
      routes,
      withComponentInputBinding(),
      withViewTransitions()
    ),
    provideClientHydration(withEventReplay()),
    provideHttpClient(
      withFetch(),
      withInterceptors([authInterceptor, namespaceInterceptor, errorInterceptor]),
      withXsrfConfiguration({
        cookieName: 'XSRF-TOKEN',
        headerName: 'X-XSRF-TOKEN',
      })
    ),
    {
      provide: APP_INITIALIZER,
      useFactory: initializeApp,
      deps: [RuntimeConfigService],
      multi: true,
    },
  ],
};
