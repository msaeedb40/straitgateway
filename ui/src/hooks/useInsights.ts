import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import { DataplaneMetrics, EbpfProgramStatus, FlowLogEntry } from '../types/observability'
import { ApiError } from '../types/api'
import { getRuntimeConfig } from '../config/runtime'

interface UseInsightsResult {
  metrics: DataplaneMetrics | null
  flows: FlowLogEntry[]
  ebpfPrograms: EbpfProgramStatus[]
  loading: boolean
  error: ApiError | null
  refetch: () => Promise<void>
}

export function useInsights(flowLimit = 50, intervalMs?: number): UseInsightsResult {
  const [metrics, setMetrics] = useState<DataplaneMetrics | null>(null)
  const [flows, setFlows] = useState<FlowLogEntry[]>([])
  const [ebpfPrograms, setEbpfPrograms] = useState<EbpfProgramStatus[]>([])
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<ApiError | null>(null)

  const pollingInterval = intervalMs ?? getRuntimeConfig().refreshIntervalMs

  const fetchData = useCallback(async (signal?: AbortSignal) => {
    try {
      const [m, f, prog] = await Promise.all([
        api.observability.getMetrics(signal),
        api.observability.getFlowLogs(flowLimit, signal),
        api.observability.getEbpfPrograms(signal),
      ])
      setMetrics(m)
      setFlows(f)
      setEbpfPrograms(prog)
      setError(null)
    } catch (err) {
      setError(err as ApiError)
    } finally {
      setLoading(false)
    }
  }, [flowLimit])

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

  return { metrics, flows, ebpfPrograms, loading, error, refetch: () => fetchData() }
}
