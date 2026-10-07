import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import { GatewayResource, TransitGatewayResource, HttpRouteResource } from '../types/gateway'
import { ApiError } from '../types/api'
import { getRuntimeConfig } from '../config/runtime'

interface UseGatewaysResult {
  gateways: GatewayResource[]
  transitGateways: TransitGatewayResource[]
  httpRoutes: HttpRouteResource[]
  loading: boolean
  error: ApiError | null
  refetch: () => Promise<void>
}

export function useGateways(namespace?: string, intervalMs?: number): UseGatewaysResult {
  const [gateways, setGateways] = useState<GatewayResource[]>([])
  const [transitGateways, setTransitGateways] = useState<TransitGatewayResource[]>([])
  const [httpRoutes, setHttpRoutes] = useState<HttpRouteResource[]>([])
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<ApiError | null>(null)

  const pollingInterval = intervalMs ?? getRuntimeConfig().refreshIntervalMs

  const fetchData = useCallback(async (signal?: AbortSignal) => {
    try {
      const [gw, tgw, routes] = await Promise.all([
        api.gateways.getGateways(namespace, signal),
        api.gateways.getTransitGateways(signal),
        api.gateways.getHttpRoutes(namespace, signal),
      ])
      setGateways(gw)
      setTransitGateways(tgw)
      setHttpRoutes(routes)
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

  return { gateways, transitGateways, httpRoutes, loading, error, refetch: () => fetchData() }
}
