import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import { ServiceInfo } from '../types/cluster'
import { ApiError } from '../types/api'
import { getRuntimeConfig } from '../config/runtime'

interface UseServicesResult {
  services: ServiceInfo[]
  loading: boolean
  error: ApiError | null
  refetch: () => Promise<void>
}

export function useServices(namespace?: string, intervalMs?: number): UseServicesResult {
  const [services, setServices] = useState<ServiceInfo[]>([])
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<ApiError | null>(null)

  const pollingInterval = intervalMs ?? getRuntimeConfig().refreshIntervalMs

  const fetchData = useCallback(async (signal?: AbortSignal) => {
    try {
      const data = await api.services.getServices(namespace, signal)
      setServices(data)
      setError(null)
    } catch (err) {
      setError(err as ApiError)
    } finally {
      setLoading(false)
    }
  }, [namespace])

  useEffect(() => {
    const controller = new AbortController()
    fetchData(controller.signal)

    if (pollingInterval > 0) {
      const timer = setInterval(() => {
        fetchData()
      }, pollingInterval)
      return () => {
        controller.abort()
        clearInterval(timer)
      }
    }
    return () => controller.abort()
  }, [fetchData, pollingInterval])

  return { services, loading, error, refetch: () => fetchData() }
}
