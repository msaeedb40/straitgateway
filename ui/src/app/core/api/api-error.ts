// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { HttpErrorResponse } from '@angular/common/http';

export interface ApiErrorDetails {
  status: number;
  statusText: string;
  message: string;
  url?: string;
  timestamp: string;
  errorPayload?: any;
}

export class ApiError extends Error {
  readonly status: number;
  readonly statusText: string;
  readonly url?: string;
  readonly timestamp: string;
  readonly rawError: any;

  constructor(httpError: HttpErrorResponse | Error | string) {
    if (httpError instanceof HttpErrorResponse) {
      const msg =
        httpError.error?.message ||
        httpError.error?.error ||
        httpError.message ||
        `Request failed with status ${httpError.status}`;
      super(msg);
      this.status = httpError.status;
      this.statusText = httpError.statusText;
      this.url = httpError.url || undefined;
      this.timestamp = new Date().toISOString();
      this.rawError = httpError.error;
    } else if (httpError instanceof Error) {
      super(httpError.message);
      this.status = 0;
      this.statusText = 'Client Error';
      this.timestamp = new Date().toISOString();
      this.rawError = httpError;
    } else {
      super(httpError);
      this.status = 0;
      this.statusText = 'Unknown';
      this.timestamp = new Date().toISOString();
      this.rawError = null;
    }
    this.name = 'ApiError';
  }

  getUserFriendlyMessage(): string {
    if (this.status === 0) {
      return 'Network error: Cannot reach the straitgateway-controller. Running in fallback mode.';
    }
    if (this.status === 401 || this.status === 403) {
      return 'Authentication failed: Insufficient permissions to access this resource.';
    }
    if (this.status === 404) {
      return 'Resource not found.';
    }
    if (this.status >= 500) {
      return `Server error (${this.status}): Controller encountered an internal failure.`;
    }
    return this.message;
  }
}
