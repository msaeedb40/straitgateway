import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import { ClusterOverview } from '../types/cluster'
import { ApiError } from '../types/api'
import { getRuntimeConfig } from '../config/runtime'

interface UseClusterResult {
  overview: ClusterOverview | null
  loading: boolean
  error: ApiError | null
  lastUpdated: Date | null
  refetch: () => Promise<void>
}

export function useCluster(intervalMs?: number): UseClusterResult {
  const [overview, setOverview] = useState<ClusterOverview | null>(null)
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<ApiError | null>(null)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)

  const pollingInterval = intervalMs ?? getRuntimeConfig().refreshIntervalMs

  const fetchData = useCallback(async (signal?: AbortSignal) => {
    try {
      const data = await api.cluster.getOverview(signal)
      setOverview(data)
      setError(null)
      setLastUpdated(new Date())
    } catch (err) {
      setError(err as ApiError)
    } finally {
      setLoading(false)
    }
  }, [])

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

  return {
    overview,
    loading,
    error,
    lastUpdated,
    refetch: () => fetchData(),
  }
}
