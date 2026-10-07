import { useState, useEffect, useCallback } from 'react'
import { api } from '../api'
import { ApiError } from '../types/api'
import { getRuntimeConfig } from '../config/runtime'

export interface TopologyNode {
  id: string
  name: string
  type: 'node' | 'pod' | 'service' | 'gateway' | 'external'
  ip?: string
  status: string
  details?: Record<string, string | number>
}

export interface TopologyLink {
  source: string
  target: string
  type: 'netkit' | 'wireguard' | 'bgp' | 'route' | 'service-lb'
  label?: string
}

export interface TopologyGraph {
  nodes: TopologyNode[]
  links: TopologyLink[]
}

interface UseTopologyResult {
  graph: TopologyGraph
  loading: boolean
  error: ApiError | null
  refetch: () => Promise<void>
}

export function useTopology(intervalMs?: number): UseTopologyResult {
  const [graph, setGraph] = useState<TopologyGraph>({ nodes: [], links: [] })
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<ApiError | null>(null)

  const pollingInterval = intervalMs ?? getRuntimeConfig().refreshIntervalMs

  const fetchData = useCallback(async (signal?: AbortSignal) => {
    try {
      const [nodes, pods, services, gws, wgPeers] = await Promise.all([
        api.nodes.getNodes(signal),
        api.pods.getPods(undefined, signal),
        api.services.getServices(undefined, signal),
        api.gateways.getGateways(undefined, signal),
        api.routes.getWireguardPeers(signal),
      ])

      const graphNodes: TopologyNode[] = []
      const graphLinks: TopologyLink[] = []

      // 1. Add K8s Nodes
      nodes.forEach((n) => {
        graphNodes.push({
          id: `node-${n.name}`,
          name: n.name,
          type: 'node',
          ip: n.ip,
          status: n.status,
          details: {
            role: n.role,
            podCIDR: n.podCIDR,
            ebpfProgs: n.ebpfLoadedCount,
          },
        })
      })

      // WireGuard inter-node links
      if (nodes.length > 1) {
        for (let i = 0; i < nodes.length - 1; i++) {
          for (let j = i + 1; j < nodes.length; j++) {
            graphLinks.push({
              source: `node-${nodes[i].name}`,
              target: `node-${nodes[j].name}`,
              type: 'wireguard',
              label: 'wg0 (mesh)',
            })
          }
        }
      }

      // 2. Add Pods & netkit links
      pods.forEach((p) => {
        const podId = `pod-${p.namespace}-${p.name}`
        graphNodes.push({
          id: podId,
          name: p.name,
          type: 'pod',
          ip: p.ip,
          status: p.status,
          details: {
            namespace: p.namespace,
            netkit: p.netkitInterface,
            identity: p.securityIdentity,
          },
        })

        if (p.nodeName) {
          graphLinks.push({
            source: `node-${p.nodeName}`,
            target: podId,
            type: 'netkit',
            label: p.netkitInterface || 'netkit0',
          })
        }
      })

      // 3. Add Services & LB mappings
      services.forEach((s) => {
        const svcId = `svc-${s.namespace}-${s.name}`
        graphNodes.push({
          id: svcId,
          name: s.name,
          type: 'service',
          ip: s.clusterIP,
          status: 'Active',
          details: {
            namespace: s.namespace,
            type: s.type,
            algorithm: s.algorithm,
          },
        })
      })

      // 4. Add Gateways
      gws.forEach((gw) => {
        const gwId = `gw-${gw.namespace}-${gw.name}`
        graphNodes.push({
          id: gwId,
          name: gw.name,
          type: 'gateway',
          status: gw.ready ? 'Ready' : 'NotReady',
          details: {
            class: gw.gatewayClassName,
            listeners: gw.listeners.length,
          },
        })
      })

      // 5. External WireGuard peers
      wgPeers.forEach((peer, idx) => {
        const peerId = `wg-peer-${idx}`
        graphNodes.push({
          id: peerId,
          name: `Peer (${peer.endpoint || peer.publicKey.substring(0, 8)})`,
          type: 'external',
          status: 'Active',
          details: {
            allowedIPs: peer.allowedIPs.join(', '),
          },
        })

        if (nodes.length > 0) {
          graphLinks.push({
            source: `node-${nodes[0].name}`,
            target: peerId,
            type: 'wireguard',
            label: 'tunnel',
          })
        }
      })

      setGraph({ nodes: graphNodes, links: graphLinks })
      setError(null)
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

  return { graph, loading, error, refetch: () => fetchData() }
}
