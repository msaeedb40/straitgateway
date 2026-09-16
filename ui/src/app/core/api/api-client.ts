import { HttpClient, HttpErrorResponse, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { ApiError } from './api-error';
import { RuntimeConfigService } from '../config/runtime-config.service';

export interface RequestOptions {
  params?: Record<string, string | number | boolean | string[]>;
  headers?: Record<string, string>;
}

@Injectable({ providedIn: 'root' })
export class ApiClient {
  private readonly http = inject(HttpClient);
  private readonly runtimeConfig = inject(RuntimeConfigService);

  get controllerBase(): string {
    return this.runtimeConfig.controllerBase;
  }

  get<T>(path: string, options?: RequestOptions): Observable<T> {
    return this.http
      .get<T>(this.resolve(path), { params: this.buildParams(options?.params) })
      .pipe(catchError(this.handleError));
  }

  post<T>(path: string, body: unknown, options?: RequestOptions): Observable<T> {
    return this.http
      .post<T>(this.resolve(path), body, { params: this.buildParams(options?.params) })
      .pipe(catchError(this.handleError));
  }

  put<T>(path: string, body: unknown, options?: RequestOptions): Observable<T> {
    return this.http
      .put<T>(this.resolve(path), body, { params: this.buildParams(options?.params) })
      .pipe(catchError(this.handleError));
  }

  patch<T>(path: string, body: unknown, options?: RequestOptions): Observable<T> {
    return this.http
      .patch<T>(this.resolve(path), body, { params: this.buildParams(options?.params) })
      .pipe(catchError(this.handleError));
  }

  delete<T>(path: string, options?: RequestOptions): Observable<T> {
    return this.http
      .delete<T>(this.resolve(path), { params: this.buildParams(options?.params) })
      .pipe(catchError(this.handleError));
  }

  /** Build a URL for an external backend (prometheus, jaeger, etc.) — not controller-relative */
  external<T>(baseUrl: string, path: string, options?: RequestOptions): Observable<T> {
    return this.http
      .get<T>(`${baseUrl}${path}`, { params: this.buildParams(options?.params) })
      .pipe(catchError(this.handleError));
  }

  private resolve(path: string): string {
    return `${this.controllerBase}${path}`;
  }

  private buildParams(
    params?: Record<string, string | number | boolean | string[]>
  ): HttpParams {
    let httpParams = new HttpParams();
    if (!params) return httpParams;
    for (const [key, value] of Object.entries(params)) {
      if (value === undefined || value === null) continue;
      if (Array.isArray(value)) {
        value.forEach((v) => (httpParams = httpParams.append(key, String(v))));
      } else {
        httpParams = httpParams.set(key, String(value));
      }
    }
    return httpParams;
  }

  private readonly handleError = (err: HttpErrorResponse): Observable<never> => {
    if (err.status === 0) {
      return throwError(() => ApiError.network(err.message));
    }
    const message: string =
      err.error?.message ?? err.error?.error ?? err.statusText ?? 'Request failed';
    const detail: string | undefined = err.error?.detail ?? err.error?.details;
    return throwError(() => ApiError.fromHttpStatus(err.status, message, detail));
  };
}
