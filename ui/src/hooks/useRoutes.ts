import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import { RouteEntry, BgpPeerSession, BfdSession, WireguardPeer, SecurityPolicyRule } from '../types/networking'
import { ApiError } from '../types/api'
import { getRuntimeConfig } from '../config/runtime'

interface UseRoutesResult {
  routes: RouteEntry[]
  bgpPeers: BgpPeerSession[]
  bfdSessions: BfdSession[]
  wireguardPeers: WireguardPeer[]
  securityPolicies: SecurityPolicyRule[]
  loading: boolean
  error: ApiError | null
  refetch: () => Promise<void>
}

export function useRoutes(vrfId?: number, intervalMs?: number): UseRoutesResult {
  const [routes, setRoutes] = useState<RouteEntry[]>([])
  const [bgpPeers, setBgpPeers] = useState<BgpPeerSession[]>([])
  const [bfdSessions, setBfdSessions] = useState<BfdSession[]>([])
  const [wireguardPeers, setWireguardPeers] = useState<WireguardPeer[]>([])
  const [securityPolicies, setSecurityPolicies] = useState<SecurityPolicyRule[]>([])
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<ApiError | null>(null)

  const pollingInterval = intervalMs ?? getRuntimeConfig().refreshIntervalMs

  const fetchData = useCallback(async (signal?: AbortSignal) => {
    try {
      const [r, bgp, bfd, wg, pol] = await Promise.all([
        api.routes.getRoutes(vrfId, signal),
        api.routes.getBgpPeers(signal),
        api.routes.getBfdSessions(signal),
        api.routes.getWireguardPeers(signal),
        api.routes.getSecurityPolicies(undefined, signal),
      ])
      setRoutes(r)
      setBgpPeers(bgp)
      setBfdSessions(bfd)
      setWireguardPeers(wg)
      setSecurityPolicies(pol)
      setError(null)
    } catch (err) {
      setError(err as ApiError)
    } finally {
      setLoading(false)
    }
  }, [vrfId])

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

  return { routes, bgpPeers, bfdSessions, wireguardPeers, securityPolicies, loading, error, refetch: () => fetchData() }
}
