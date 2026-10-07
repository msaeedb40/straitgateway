import React from 'react'
import { Card } from '../../components/common/Card'
import { Table, Column } from '../../components/common/Table'
import { Badge } from '../../components/common/Badge'
import { useRoutes } from '../../hooks/useRoutes'
import { SecurityPolicyRule } from '../../types/networking'
import { Lock, CheckCircle2 } from 'lucide-react'

export const SecurityPoliciesView: React.FC = () => {
  const { securityPolicies, loading, refetch } = useRoutes()

  const columns: Column<SecurityPolicyRule>[] = [
    {
      header: 'Policy Name',
      accessor: (p) => (
        <div className="flex flex-col">
          <span className="text-white font-bold">{p.name}</span>
          <span className="text-[10px] text-gray-500 font-mono">{p.namespace}</span>
        </div>
      ),
    },
    {
      header: 'Direction',
      accessor: (p) => (
        <Badge variant={p.direction === 'Ingress' ? 'cyan' : 'blue'}>
          {p.direction}
        </Badge>
      ),
    },
    {
      header: 'Action',
      accessor: (p) => (
        <Badge variant={p.action === 'Allow' ? 'emerald' : 'rose'}>
          {p.action}
        </Badge>
      ),
    },
    {
      header: 'Source Selector',
      accessor: (p) => (
        <span className="text-gray-300 font-mono text-xs">{p.sourceSelector}</span>
      ),
    },
    {
      header: 'Destination Selector',
      accessor: (p) => (
        <span className="text-cyan-300 font-mono text-xs">{p.destinationSelector}</span>
      ),
    },
    {
      header: 'Ports & Protocol',
      accessor: (p) => (
        <span className="text-gray-400 font-mono text-xs">
          {p.protocol}/{p.ports.join(', ')}
        </span>
      ),
    },
    {
      header: 'eBPF Hits',
      accessor: (p) => (
        <span className="text-emerald-400 font-mono font-bold">
          {new Intl.NumberFormat().format(p.hitsCount)}
        </span>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      {/* Zero Trust Principles Banner */}
      <div className="glass-panel p-5 rounded-xl border border-emerald-500/30 bg-gradient-to-r from-emerald-950/20 via-gray-900/40 to-cyan-950/20 flex flex-wrap items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-400">
            <Lock className="w-5 h-5" />
          </div>
          <div>
            <h3 className="text-sm font-bold text-white">In-Kernel Zero-Trust Identity Enforcement</h3>
            <p className="text-xs text-gray-400 font-mono">
              Default-Deny baseline • Evaluated at XDP and TC ingress before packet routing
            </p>
          </div>
        </div>

        <div className="flex items-center gap-3 text-xs font-mono">
          <div className="flex items-center gap-1.5 text-emerald-400">
            <CheckCircle2 className="w-4 h-4" />
            <span>Policy Maps Active</span>
          </div>
          <span className="text-gray-600">|</span>
          <span className="text-gray-300">IdentityMap: 32-bit SPIFFE IDs</span>
        </div>
      </div>

      <Card
        title="Active SecurityPolicy Rules"
        subtitle="Identity-based authorization rules compiled into PolicyMap"
        action={
          <button
            onClick={() => refetch()}
            className="text-xs text-cyan-400 hover:text-cyan-300 font-mono"
          >
            {loading ? 'Refreshing...' : 'Sync Policies'}
          </button>
        }
      >
        <Table<SecurityPolicyRule>
          columns={columns}
          data={securityPolicies}
          keyExtractor={(item) => item.id}
          emptyMessage="No SecurityPolicy rules configured in cluster"
          id="table-security-policies"
        />
      </Card>
    </div>
  )
}
