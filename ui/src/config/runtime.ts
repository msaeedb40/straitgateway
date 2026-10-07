export interface RuntimeConfig {
  apiBaseUrl: string
  refreshIntervalMs: number
  clusterName: string
  environment: string
  wsBaseUrl?: string
}

declare global {
  interface Window {
    __STRAITGATEWAY_CONFIG__?: Partial<RuntimeConfig>
  }
}

/**
 * Loads runtime configuration parameterization dynamically.
 * Priority: Window runtime injection (Kubernetes ConfigMap) > LocalStorage override > Default
 */
export function getRuntimeConfig(): RuntimeConfig {
  const injected = window.__STRAITGATEWAY_CONFIG__ || {}
  const localApiUrl = localStorage.getItem('sg_api_url')
  const localRefresh = localStorage.getItem('sg_refresh_interval')

  return {
    apiBaseUrl: localApiUrl || injected.apiBaseUrl || '/api/v1',
    refreshIntervalMs: localRefresh ? parseInt(localRefresh, 10) : (injected.refreshIntervalMs || 5000),
    clusterName: injected.clusterName || 'straitgateway-prod-01',
    environment: injected.environment || 'production',
    wsBaseUrl: injected.wsBaseUrl || `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/stream`,
  }
}

export function updateRuntimeConfig(updates: Partial<RuntimeConfig>): void {
  if (updates.apiBaseUrl) {
    localStorage.setItem('sg_api_url', updates.apiBaseUrl)
  }
  if (updates.refreshIntervalMs) {
    localStorage.setItem('sg_refresh_interval', updates.refreshIntervalMs.toString())
  }
}
