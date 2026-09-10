// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, throwError } from 'rxjs';
import { ApiError } from '../api/api-error';
import { NotificationService } from '../services/notification.service';

export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const notif = inject(NotificationService, { optional: true });

  return next(req).pipe(
    catchError((err) => {
      const apiErr = new ApiError(err);
      if (apiErr.status >= 500 && notif) {
        notif.danger('API Server Error', apiErr.getUserFriendlyMessage());
      }
      return throwError(() => apiErr);
    })
  );
};
