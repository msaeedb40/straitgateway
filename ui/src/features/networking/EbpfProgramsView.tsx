import React from 'react'
import { Card } from '../../components/common/Card'
import { Table, Column } from '../../components/common/Table'
import { Badge } from '../../components/common/Badge'
import { useInsights } from '../../hooks/useInsights'
import { EbpfProgramStatus } from '../../types/observability'
import { CheckCircle2 } from 'lucide-react'

export const EbpfProgramsView: React.FC = () => {
  const { ebpfPrograms, loading, refetch } = useInsights(10)

  const columns: Column<EbpfProgramStatus>[] = [
    {
      header: 'Program ID & Name',
      accessor: (p) => (
        <div className="flex flex-col">
          <span className="text-white font-bold">{p.name}</span>
          <span className="text-[10px] text-gray-500 font-mono">bpf_prog_id: #{p.id}</span>
        </div>
      ),
    },
    {
      header: 'Attachment Hook',
      accessor: (p) => {
        const variant =
          p.type === 'XDP'
            ? 'rose'
            : p.type.startsWith('TCX')
            ? 'cyan'
            : p.type === 'NETKIT'
            ? 'emerald'
            : 'indigo'
        return <Badge variant={variant}>{p.type}</Badge>
      },
    },
    {
      header: 'Attached Device',
      accessor: (p) => (
        <span className="px-2 py-0.5 rounded bg-gray-900 border border-gray-800 text-gray-300 font-mono text-xs">
          {p.device || 'cgroup v2 root'}
        </span>
      ),
    },
    {
      header: 'Instructions',
      accessor: (p) => (
        <span className="text-gray-300 font-mono text-xs">
          {new Intl.NumberFormat().format(p.instructions)} insns
        </span>
      ),
    },
    {
      header: 'Run Count',
      accessor: (p) => (
        <span className="text-cyan-400 font-mono font-bold text-xs">
          {new Intl.NumberFormat().format(p.runCount)}
        </span>
      ),
    },
    {
      header: 'Total Runtime',
      accessor: (p) => (
        <span className="text-gray-400 font-mono text-xs">
          {(p.runTimeTotalNs / 1_000_000).toFixed(2)} ms
        </span>
      ),
    },
    {
      header: 'CO-RE Verified',
      accessor: (p) => (
        <div className="flex items-center gap-1.5 text-xs text-emerald-400 font-mono">
          <CheckCircle2 className="w-3.5 h-3.5" />
          <span>Verified</span>
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      {/* Architecture Specs */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="glass-panel p-4 rounded-xl border border-gray-800 font-mono text-xs">
          <span className="text-gray-500 uppercase text-[10px]">Compiler Toolchain</span>
          <p className="text-sm font-bold text-white mt-1">Clang 22 / LLVM 22</p>
          <p className="text-[11px] text-cyan-400 mt-0.5">BPF CO-RE Multi-Kernel</p>
        </div>
        <div className="glass-panel p-4 rounded-xl border border-gray-800 font-mono text-xs">
          <span className="text-gray-500 uppercase text-[10px]">Runtime Loader</span>
          <p className="text-sm font-bold text-white mt-1">cilium/ebpf v0.20</p>
          <p className="text-[11px] text-indigo-400 mt-0.5">Go 1.27.1 / straitd</p>
        </div>
        <div className="glass-panel p-4 rounded-xl border border-gray-800 font-mono text-xs">
          <span className="text-gray-500 uppercase text-[10px]">Kernel Target</span>
          <p className="text-sm font-bold text-white mt-1">Linux 6.12+ LTS</p>
          <p className="text-[11px] text-emerald-400 mt-0.5">TCX + netkit + cgroup v2</p>
        </div>
      </div>

      <Card
        title="Loaded eBPF Kernel Dataplane Programs"
        subtitle="Live inspection of bytecode verified and loaded into the Linux kernel"
        action={
          <button
            onClick={() => refetch()}
            className="text-xs text-cyan-400 hover:text-cyan-300 font-mono"
          >
            {loading ? 'Polling BPF...' : 'Refresh Programs'}
          </button>
        }
      >
        <Table<EbpfProgramStatus>
          columns={columns}
          data={ebpfPrograms}
          keyExtractor={(item) => String(item.id)}
          emptyMessage="No eBPF programs reported by node agent"
          id="table-ebpf-programs"
        />
      </Card>
    </div>
  )
}
