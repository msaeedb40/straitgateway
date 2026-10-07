import { ApiError, ApiResponse } from '../types/api'

export interface RequestOptions extends RequestInit {
  params?: Record<string, string | number | boolean | undefined>
  timeoutMs?: number
}

/**
 * StraitGatewayApiClient encapsulates HTTP requests using the native Fetch API.
 * Follows zero-third-party-library requirements (No Axios, No TanStack Query).
 */
export class StraitGatewayApiClient {
  private readonly baseUrl: string

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl.replace(/\/+$/, '')
  }

  public getBaseUrl(): string {
    return this.baseUrl
  }

  private buildUrl(endpoint: string, params?: Record<string, string | number | boolean | undefined>): string {
    const cleanedEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`
    const url = new URL(`${this.baseUrl}${cleanedEndpoint}`, window.location.origin)

    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          url.searchParams.append(key, String(value))
        }
      })
    }

    return url.toString()
  }

  public async request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
    const { params, timeoutMs = 8000, ...fetchOptions } = options
    const url = this.buildUrl(endpoint, params)

    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), timeoutMs)

    // Merge external signal if passed
    if (fetchOptions.signal) {
      fetchOptions.signal.addEventListener('abort', () => controller.abort())
    }

    try {
      const response = await fetch(url, {
        ...fetchOptions,
        signal: controller.signal,
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
          ...fetchOptions.headers,
        },
      })

      clearTimeout(timer)

      if (!response.ok) {
        let errDetails: Record<string, unknown> | undefined
        try {
          errDetails = await response.json()
        } catch {
          // Response body was not JSON
        }

        const error: ApiError = {
          status: response.status,
          message: `StraitGateway API Error (${response.status}): ${response.statusText}`,
          details: errDetails,
        }
        throw error
      }

      const payload = (await response.json()) as ApiResponse<T> | T
      // Check if wrapped in ApiResponse
      if (payload && typeof payload === 'object' && 'data' in payload) {
        return (payload as ApiResponse<T>).data
      }
      return payload as T
    } catch (err: unknown) {
      clearTimeout(timer)
      if (err instanceof DOMException && err.name === 'AbortError') {
        throw {
          status: 408,
          message: `Request to ${endpoint} timed out after ${timeoutMs}ms`,
        } as ApiError
      }
      if ((err as ApiError).status) {
        throw err as ApiError
      }
      throw {
        status: 0,
        message: err instanceof Error ? err.message : 'Network error or backend unreachable',
      } as ApiError
    }
  }

  public get<T>(endpoint: string, params?: Record<string, string | number | boolean | undefined>, signal?: AbortSignal): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET', params, signal })
  }

  public post<T>(endpoint: string, body?: unknown, signal?: AbortSignal): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
      signal,
    })
  }

  public delete<T>(endpoint: string, signal?: AbortSignal): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE', signal })
  }
}
