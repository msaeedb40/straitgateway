import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { NotificationService } from '../services/notification.service';
import { ApiError } from '../api/api-error';

export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const notifications = inject(NotificationService);

  return next(req).pipe(
    catchError((err: unknown) => {
      if (err instanceof HttpErrorResponse) {
        // 5xx errors trigger a global notification
        if (err.status >= 500) {
          const message =
            err.error?.message ?? err.error?.error ?? 'An unexpected server error occurred';
          notifications.error(`Server Error (${err.status})`, message);
        }
        // 503 specifically — backend unavailable
        if (err.status === 503) {
          notifications.warning('Backend Unavailable', err.url ?? undefined);
        }
      }
      // Always re-throw — components handle their own error state
      return throwError(() => err);
    })
  );
};
