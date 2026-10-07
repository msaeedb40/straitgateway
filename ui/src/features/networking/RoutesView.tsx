import React, { useState } from 'react'
import { Card } from '../../components/common/Card'
import { Table, Column } from '../../components/common/Table'
import { Badge } from '../../components/common/Badge'
import { useRoutes } from '../../hooks/useRoutes'
import { RouteEntry } from '../../types/networking'
import { Search, Layers, ArrowRight } from 'lucide-react'

export const RoutesView: React.FC = () => {
  const { routes, loading, refetch } = useRoutes()
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedVrf, setSelectedVrf] = useState<number | 'ALL'>('ALL')

  const vrfs = Array.from(new Set(routes.map((r) => r.vrfId))).sort((a, b) => a - b)

  const filteredRoutes = routes.filter((r) => {
    const matchesSearch =
      r.destinationCIDR.includes(searchTerm) ||
      r.nextHop.includes(searchTerm) ||
      r.outputInterface.includes(searchTerm)
    const matchesVrf = selectedVrf === 'ALL' || r.vrfId === selectedVrf
    return matchesSearch && matchesVrf
  })

  const routeTypeBadges: Record<string, 'cyan' | 'blue' | 'indigo' | 'emerald' | 'amber'> = {
    Local: 'cyan',
    Cluster: 'blue',
    WireGuard: 'indigo',
    BGP: 'emerald',
    Static: 'amber',
  }

  const columns: Column<RouteEntry>[] = [
    {
      header: 'Destination CIDR',
      accessor: (r) => (
        <span className="text-white font-bold tracking-wide">{r.destinationCIDR}</span>
      ),
    },
    {
      header: 'Prefix',
      accessor: (r) => <span className="text-gray-400">/{r.prefixLen}</span>,
    },
    {
      header: 'Next Hop',
      accessor: (r) => (
        <div className="flex items-center gap-1.5 text-cyan-300">
          <ArrowRight className="w-3.5 h-3.5 text-gray-500" />
          <span>{r.nextHop}</span>
        </div>
      ),
    },
    {
      header: 'Interface',
      accessor: (r) => (
        <span className="px-2 py-0.5 rounded bg-gray-900 border border-gray-800 text-gray-300 text-[11px]">
          {r.outputInterface}
        </span>
      ),
    },
    {
      header: 'Route Type',
      accessor: (r) => (
        <Badge variant={routeTypeBadges[r.routeType] || 'cyan'}>{r.routeType}</Badge>
      ),
    },
    {
      header: 'VRF ID',
      accessor: (r) => (
        <span className="text-gray-400 font-mono">VRF-{r.vrfId}</span>
      ),
    },
    {
      header: 'Metric',
      accessor: (r) => <span className="text-gray-400 font-mono">{r.metric}</span>,
    },
    {
      header: 'State',
      accessor: (r) => (
        <span
          className={`text-[11px] font-bold ${
            r.status === 'Active' ? 'text-emerald-400' : 'text-gray-500'
          }`}
        >
          {r.status}
        </span>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      {/* Search and VRF Filter Bar */}
      <div className="glass-panel p-4 rounded-xl border border-gray-800 flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-3 flex-1 min-w-[240px]">
          <div className="relative w-full max-w-sm">
            <Search className="w-4 h-4 text-gray-500 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              id="input-routes-search"
              placeholder="Search destination prefix, next hop..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-9 pr-3 py-1.5 bg-gray-900 border border-gray-800 rounded-lg text-xs text-gray-200 font-mono focus:outline-none focus:border-cyan-500"
            />
          </div>

          {/* VRF Domain Switcher */}
          <div className="flex items-center gap-1.5 text-xs font-mono">
            <Layers className="w-3.5 h-3.5 text-gray-500" />
            <button
              onClick={() => setSelectedVrf('ALL')}
              className={`px-2.5 py-1 rounded-lg text-[10px] font-bold tracking-wider transition-colors ${
                selectedVrf === 'ALL'
                  ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40'
                  : 'text-gray-400 hover:text-gray-200 bg-gray-900/60'
              }`}
            >
              ALL VRFS
            </button>
            {vrfs.map((vrf) => (
              <button
                key={vrf}
                onClick={() => setSelectedVrf(vrf)}
                className={`px-2.5 py-1 rounded-lg text-[10px] font-bold tracking-wider transition-colors ${
                  selectedVrf === vrf
                    ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40'
                    : 'text-gray-400 hover:text-gray-200 bg-gray-900/60'
                }`}
              >
                VRF-{vrf}
              </button>
            ))}
          </div>
        </div>

        <button
          onClick={() => refetch()}
          className="text-xs text-cyan-400 hover:text-cyan-300 font-mono"
        >
          {loading ? 'Refreshing FIB...' : 'Sync FIB'}
        </button>
      </div>

      <Card
        title="Forwarding Information Base (eBPF FIB)"
        subtitle={`Showing ${filteredRoutes.length} active routes synchronized to LPM trie (RouteMap)`}
      >
        <Table<RouteEntry>
          columns={columns}
          data={filteredRoutes}
          keyExtractor={(item, idx) => `${item.destinationCIDR}-${item.vrfId}-${idx}`}
          emptyMessage="No routes found in the eBPF forwarding table"
          id="table-routes"
        />
      </Card>
    </div>
  )
}
