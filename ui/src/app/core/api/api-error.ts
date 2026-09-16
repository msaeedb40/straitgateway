export type ApiErrorCode =
  | 'NETWORK_ERROR'
  | 'UNAUTHORIZED'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'CONFLICT'
  | 'VALIDATION_ERROR'
  | 'SERVER_ERROR'
  | 'BACKEND_UNAVAILABLE'
  | 'UNKNOWN';

export class ApiError extends Error {
  readonly code: ApiErrorCode;
  readonly status: number | null;
  readonly detail: string | null;
  readonly retryable: boolean;

  constructor(params: {
    code: ApiErrorCode;
    message: string;
    status?: number;
    detail?: string;
    retryable?: boolean;
  }) {
    super(params.message);
    this.name = 'ApiError';
    this.code = params.code;
    this.status = params.status ?? null;
    this.detail = params.detail ?? null;
    this.retryable = params.retryable ?? false;
  }

  static fromHttpStatus(status: number, message: string, detail?: string): ApiError {
    switch (true) {
      case status === 401:
        return new ApiError({ code: 'UNAUTHORIZED', status, message, detail, retryable: false });
      case status === 403:
        return new ApiError({ code: 'FORBIDDEN', status, message, detail, retryable: false });
      case status === 404:
        return new ApiError({ code: 'NOT_FOUND', status, message, detail, retryable: false });
      case status === 409:
        return new ApiError({ code: 'CONFLICT', status, message, detail, retryable: false });
      case status === 422:
        return new ApiError({ code: 'VALIDATION_ERROR', status, message, detail, retryable: false });
      case status === 503:
        return new ApiError({ code: 'BACKEND_UNAVAILABLE', status, message, detail, retryable: true });
      case status >= 500:
        return new ApiError({ code: 'SERVER_ERROR', status, message, detail, retryable: true });
      default:
        return new ApiError({ code: 'UNKNOWN', status, message, detail, retryable: false });
    }
  }

  static network(detail?: string): ApiError {
    return new ApiError({
      code: 'NETWORK_ERROR',
      message: 'Network request failed',
      detail,
      retryable: true,
    });
  }
}
