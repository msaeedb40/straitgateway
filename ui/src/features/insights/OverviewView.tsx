import React from 'react'
import {
  Server,
  Boxes,
  Layers,
  Radio,
  Lock,
  Zap,
  Activity,
  AlertTriangle,
  ArrowUpRight,
} from 'lucide-react'
import { MetricCard } from '../../components/common/MetricCard'
import { Card } from '../../components/common/Card'
import { StatusDot } from '../../components/common/StatusDot'
import { Badge } from '../../components/common/Badge'
import { useCluster } from '../../hooks/useCluster'
import { useInsights } from '../../hooks/useInsights'

interface OverviewViewProps {
  onNavigate: (tab: string) => void
}

export const OverviewView: React.FC<OverviewViewProps> = ({ onNavigate }) => {
  const { overview, error: clusterError } = useCluster()
  const { metrics, ebpfPrograms, loading: insightsLoading } = useInsights(10)

  if (clusterError) {
    return (
      <div className="p-6">
        <div className="glass-panel rounded-xl p-8 border border-rose-500/30 text-center">
          <AlertTriangle className="w-12 h-12 text-rose-400 mx-auto mb-3" />
          <h3 className="text-lg font-bold text-white mb-2">Connecting to StraitGateway Daemon</h3>
          <p className="text-sm text-gray-400 max-w-lg mx-auto mb-4">
            Could not reach the StraitGateway API at the configured endpoint.
            Verify that <code className="text-cyan-400">straitd</code> or <code className="text-cyan-400">sg-controller</code> is running.
          </p>
          <div className="text-xs font-mono bg-gray-950 p-3 rounded-lg border border-gray-800 text-rose-300 max-w-md mx-auto">
            {clusterError.message}
          </div>
        </div>
      </div>
    )
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const formatNumber = (num: number) => {
    return new Intl.NumberFormat().format(num)
  }

  return (
    <div className="space-y-6">
      {/* Top Banner / Cluster Status */}
      <div className="glass-panel rounded-xl p-5 border border-cyan-500/20 bg-gradient-to-r from-cyan-950/20 via-gray-900/40 to-indigo-950/20 flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="p-3 rounded-xl bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
            <Zap className="w-6 h-6 animate-pulse" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-base font-bold text-white tracking-tight">
                {overview?.name || 'StraitGateway Transit Mesh'}
              </h2>
              <StatusDot status={overview?.status || 'Active'} pulse />
            </div>
            <p className="text-xs text-gray-400 font-mono mt-0.5">
              Kernel {overview?.kernelVersion || '6.12.0-sg'} • eBPF CO-RE • netkit Dataplane
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2 font-mono text-xs">
          <span className="text-gray-400">K8s:</span>
          <span className="text-gray-200 font-bold">{overview?.kubernetesVersion || 'v1.32.2'}</span>
          <span className="text-gray-600">|</span>
          <span className="text-gray-400">Uptime:</span>
          <span className="text-cyan-400 font-bold">
            {overview?.uptimeSeconds ? `${Math.floor(overview.uptimeSeconds / 3600)}h` : '99.99%'}
          </span>
        </div>
      </div>

      {/* Primary KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard
          label="Cluster Nodes"
          value={overview?.nodesCount ?? 0}
          unit="Nodes"
          change="All eBPF netkit ready"
          trend="up"
          variant="cyan"
          icon={<Server className="w-5 h-5 text-cyan-400" />}
        />
        <MetricCard
          label="Active Pods"
          value={overview?.podsCount ?? 0}
          unit="netkit workloads"
          change="Zero-veth overhead"
          trend="up"
          variant="blue"
          icon={<Boxes className="w-5 h-5 text-blue-400" />}
        />
        <MetricCard
          label="WireGuard Mesh"
          value={overview?.activeWireguardPeers ?? 0}
          unit="Peers"
          change="ChaCha20-Poly1305 encrypted"
          trend="up"
          variant="indigo"
          icon={<Lock className="w-5 h-5 text-indigo-400" />}
        />
        <MetricCard
          label="BGP Peering"
          value={overview?.bgpSessionsUp ?? 0}
          unit="Sessions"
          change="BFD sub-second failover"
          trend="up"
          variant="emerald"
          icon={<Radio className="w-5 h-5 text-emerald-400" />}
        />
      </div>

      {/* Dataplane Real-Time Performance */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card
          title="Dataplane Throughput"
          subtitle="Real-time kernel eBPF packet counters"
          className="lg:col-span-2"
          action={
            <button
              onClick={() => onNavigate('flows')}
              className="text-xs text-cyan-400 hover:text-cyan-300 flex items-center gap-1 font-mono"
            >
              View Flows <ArrowUpRight className="w-3.5 h-3.5" />
            </button>
          }
        >
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 py-2 font-mono">
            <div className="p-3 rounded-lg bg-gray-900/60 border border-gray-800">
              <span className="text-[11px] text-gray-500 uppercase">Packets Total</span>
              <p className="text-xl font-bold text-white mt-1">
                {formatNumber(metrics?.packetsTotal || 0)}
              </p>
            </div>
            <div className="p-3 rounded-lg bg-gray-900/60 border border-gray-800">
              <span className="text-[11px] text-gray-500 uppercase">Bytes Forwarded</span>
              <p className="text-xl font-bold text-cyan-400 mt-1">
                {formatBytes(metrics?.bytesTotal || 0)}
              </p>
            </div>
            <div className="p-3 rounded-lg bg-gray-900/60 border border-gray-800">
              <span className="text-[11px] text-gray-500 uppercase">Conntrack Flows</span>
              <p className="text-xl font-bold text-indigo-400 mt-1">
                {formatNumber(metrics?.conntrackEntries || 0)}
              </p>
            </div>
            <div className="p-3 rounded-lg bg-gray-900/60 border border-gray-800">
              <span className="text-[11px] text-gray-500 uppercase">Drops & Policy Denied</span>
              <p className="text-xl font-bold text-rose-400 mt-1">
                {formatNumber((metrics?.dropsTotal || 0) + (metrics?.policyDeniedTotal || 0))}
              </p>
            </div>
          </div>

          <div className="mt-4 pt-4 border-t border-gray-800/80 flex items-center justify-between text-xs text-gray-400">
            <div className="flex items-center gap-2">
              <Activity className="w-4 h-4 text-emerald-400" />
              <span>eBPF Maglev LB Requests:</span>
              <span className="text-white font-mono font-bold">
                {formatNumber(metrics?.serviceLbRequests || 0)}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <span>WireGuard I/O:</span>
              <span className="text-cyan-300 font-mono">
                RX {formatBytes(metrics?.wireguardRxBytes || 0)} / TX {formatBytes(metrics?.wireguardTxBytes || 0)}
              </span>
            </div>
          </div>
        </Card>

        {/* eBPF Programs Status */}
        <Card
          title="Active eBPF Programs"
          subtitle="Loaded hooks & verifier state"
          action={
            <button
              onClick={() => onNavigate('ebpf')}
              className="text-xs text-cyan-400 hover:text-cyan-300 flex items-center gap-1 font-mono"
            >
              Details <ArrowUpRight className="w-3.5 h-3.5" />
            </button>
          }
        >
          <div className="space-y-3">
            {ebpfPrograms.length === 0 ? (
              <div className="text-center py-6 text-xs text-gray-500">
                {insightsLoading ? 'Querying BPF subsystem...' : 'No programs registered yet'}
              </div>
            ) : (
              ebpfPrograms.slice(0, 5).map((prog) => (
                <div
                  key={prog.id}
                  className="flex items-center justify-between p-2.5 rounded-lg bg-gray-900/50 border border-gray-800/60 font-mono text-xs"
                >
                  <div>
                    <p className="text-gray-200 font-semibold">{prog.name}</p>
                    <p className="text-[11px] text-gray-500">
                      {prog.device || 'cgroup'} • {prog.instructions} insns
                    </p>
                  </div>
                  <Badge variant="cyan">{prog.type}</Badge>
                </div>
              ))
            )}
          </div>
        </Card>
      </div>

      {/* Quick Navigation Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <button
          onClick={() => onNavigate('topology')}
          className="glass-panel p-4 rounded-xl border border-gray-800 hover:border-cyan-500/50 text-left transition-all group"
        >
          <Layers className="w-5 h-5 text-cyan-400 mb-2 group-hover:scale-110 transition-transform" />
          <h4 className="text-sm font-semibold text-white">Interactive Topology</h4>
          <p className="text-xs text-gray-400 mt-1">
            Explore live D3 force graph of nodes, pod netkit links, and WireGuard tunnels.
          </p>
        </button>

        <button
          onClick={() => onNavigate('routing-bgp')}
          className="glass-panel p-4 rounded-xl border border-gray-800 hover:border-indigo-500/50 text-left transition-all group"
        >
          <Radio className="w-5 h-5 text-indigo-400 mb-2 group-hover:scale-110 transition-transform" />
          <h4 className="text-sm font-semibold text-white">Transit Routing & BGP</h4>
          <p className="text-xs text-gray-400 mt-1">
            Inspect BGP route advertisements, BFD session heartbeats, and FIB tables.
          </p>
        </button>

        <button
          onClick={() => onNavigate('security')}
          className="glass-panel p-4 rounded-xl border border-gray-800 hover:border-emerald-500/50 text-left transition-all group"
        >
          <Lock className="w-5 h-5 text-emerald-400 mb-2 group-hover:scale-110 transition-transform" />
          <h4 className="text-sm font-semibold text-white">Zero-Trust Policies</h4>
          <p className="text-xs text-gray-400 mt-1">
            Manage eBPF-enforced identity policies, default-deny rules, and audit violations.
          </p>
        </button>
      </div>
    </div>
  )
}
