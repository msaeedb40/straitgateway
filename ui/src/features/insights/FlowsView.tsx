import React, { useState } from 'react'
import { Card } from '../../components/common/Card'
import { Table, Column } from '../../components/common/Table'
import { Badge } from '../../components/common/Badge'
import { useInsights } from '../../hooks/useInsights'
import { FlowLogEntry } from '../../types/observability'
import { Search, Activity, Filter } from 'lucide-react'

export const FlowsView: React.FC = () => {
  const { flows, loading, refetch } = useInsights(100)
  const [searchTerm, setSearchTerm] = useState('')
  const [actionFilter, setActionFilter] = useState<string>('ALL')

  const filteredFlows = flows.filter((f) => {
    const matchesSearch =
      f.sourceIp.includes(searchTerm) ||
      f.destinationIp.includes(searchTerm) ||
      (f.ruleName && f.ruleName.toLowerCase().includes(searchTerm.toLowerCase()))
    const matchesAction = actionFilter === 'ALL' || f.action === actionFilter
    return matchesSearch && matchesAction
  })

  const columns: Column<FlowLogEntry>[] = [
    {
      header: 'Timestamp',
      accessor: (f) => (
        <span className="text-gray-400">
          {new Date(f.timestamp).toLocaleTimeString()}
        </span>
      ),
    },
    {
      header: 'Source',
      accessor: (f) => (
        <div className="flex flex-col">
          <span className="text-white font-bold">{f.sourceIp}:{f.sourcePort}</span>
          <span className="text-[10px] text-gray-500">Identity: #{f.sourceIdentity}</span>
        </div>
      ),
    },
    {
      header: 'Destination',
      accessor: (f) => (
        <div className="flex flex-col">
          <span className="text-cyan-300 font-bold">{f.destinationIp}:{f.destinationPort}</span>
          <span className="text-[10px] text-gray-500">Identity: #{f.destinationIdentity}</span>
        </div>
      ),
    },
    {
      header: 'Proto',
      accessor: (f) => <Badge variant="blue">{f.protocol}</Badge>,
    },
    {
      header: 'Action',
      accessor: (f) => (
        <Badge
          variant={f.action === 'ALLOW' ? 'emerald' : f.action === 'DENY' ? 'rose' : 'amber'}
        >
          {f.action}
        </Badge>
      ),
    },
    {
      header: 'Interface / Node',
      accessor: (f) => (
        <div className="flex flex-col text-[11px]">
          <span className="text-gray-300">{f.interface}</span>
          <span className="text-gray-500">{f.node}</span>
        </div>
      ),
    },
    {
      header: 'Packets / Bytes',
      accessor: (f) => (
        <span className="text-gray-400">
          {f.packets} pkts • {(f.bytes / 1024).toFixed(1)} KB
        </span>
      ),
    },
    {
      header: 'Policy Rule',
      accessor: (f) => (
        <span className="text-gray-400 text-[11px] truncate max-w-[120px] block">
          {f.ruleName || 'default-allow'}
        </span>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      {/* Filter and Search Bar */}
      <div className="glass-panel p-4 rounded-xl border border-gray-800 flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-3 flex-1 min-w-[240px]">
          <div className="relative w-full max-w-sm">
            <Search className="w-4 h-4 text-gray-500 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              id="input-flows-search"
              placeholder="Search IP, destination, or policy rule..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-9 pr-3 py-1.5 bg-gray-900 border border-gray-800 rounded-lg text-xs text-gray-200 font-mono focus:outline-none focus:border-cyan-500"
            />
          </div>

          <div className="flex items-center gap-1.5 text-xs font-mono">
            <Filter className="w-3.5 h-3.5 text-gray-500" />
            {['ALL', 'ALLOW', 'DENY', 'REDIRECT'].map((act) => (
              <button
                key={act}
                onClick={() => setActionFilter(act)}
                className={`px-2.5 py-1 rounded-lg text-[10px] font-bold tracking-wider transition-colors ${
                  actionFilter === act
                    ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40'
                    : 'text-gray-400 hover:text-gray-200 bg-gray-900/60'
                }`}
              >
                {act}
              </button>
            ))}
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => refetch()}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-gray-900 hover:bg-gray-800 text-gray-300 border border-gray-800 rounded-lg text-xs font-mono transition-colors"
          >
            <Activity className="w-3.5 h-3.5 text-cyan-400 animate-pulse" />
            <span>{loading ? 'Polling RingBuffer...' : 'Poll Stream'}</span>
          </button>
        </div>
      </div>

      {/* Flows Table */}
      <Card
        title="Kernel Ring Buffer Flow Telemetry"
        subtitle={`Showing ${filteredFlows.length} flows intercepted at eBPF hooks`}
      >
        <Table<FlowLogEntry>
          columns={columns}
          data={filteredFlows}
          keyExtractor={(item) => item.id}
          emptyMessage="No flows matching filter criteria recorded in ring buffer"
          id="table-flows"
        />
      </Card>
    </div>
  )
}
