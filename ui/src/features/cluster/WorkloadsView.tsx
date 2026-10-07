import React, { useState } from 'react'
import { Card } from '../../components/common/Card'
import { Table, Column } from '../../components/common/Table'
import { Badge } from '../../components/common/Badge'
import { StatusDot } from '../../components/common/StatusDot'
import { useNodes } from '../../hooks/useNodes'
import { usePods } from '../../hooks/usePods'
import { useServices } from '../../hooks/useServices'
import { NodeInfo, PodInfo, ServiceInfo } from '../../types/cluster'
import { Server, Boxes, Layers, CheckCircle2 } from 'lucide-react'

export const WorkloadsView: React.FC = () => {
  const [subTab, setSubTab] = useState<'nodes' | 'pods' | 'services'>('nodes')
  const { nodes, loading: nodesLoading, refetch: refetchNodes } = useNodes()
  const { pods, loading: podsLoading, refetch: refetchPods } = usePods()
  const { services, loading: servicesLoading, refetch: refetchServices } = useServices()

  const nodeColumns: Column<NodeInfo>[] = [
    {
      header: 'Node Name',
      accessor: (n) => <span className="text-white font-bold">{n.name}</span>,
    },
    {
      header: 'Internal IP',
      accessor: (n) => <span className="text-cyan-400 font-mono text-xs">{n.ip}</span>,
    },
    {
      header: 'Pod CIDR',
      accessor: (n) => <span className="text-gray-300 font-mono text-xs">{n.podCIDR}</span>,
    },
    {
      header: 'Kernel / Dataplane',
      accessor: (n) => (
        <span className="text-gray-400 font-mono text-[11px]">
          {n.kernelVersion} • netkit:{n.netkitSupported ? 'yes' : 'no'}
        </span>
      ),
    },
    {
      header: 'eBPF Programs',
      accessor: (n) => (
        <Badge variant="cyan">{n.ebpfLoadedCount} programs</Badge>
      ),
    },
    {
      header: 'WireGuard Endpoint',
      accessor: (n) => (
        <span className="text-indigo-400 font-mono text-[11px]">
          {n.wireguardEndpoint || 'wg0'}
        </span>
      ),
    },
    {
      header: 'Status',
      accessor: (n) => <StatusDot status={n.status} pulse={n.status === 'Ready'} />,
    },
  ]

  const podColumns: Column<PodInfo>[] = [
    {
      header: 'Pod Name',
      accessor: (p) => (
        <div className="flex flex-col">
          <span className="text-white font-bold">{p.name}</span>
          <span className="text-[10px] text-gray-500 font-mono">ns: {p.namespace}</span>
        </div>
      ),
    },
    {
      header: 'Pod IP',
      accessor: (p) => <span className="text-cyan-400 font-mono text-xs">{p.ip}</span>,
    },
    {
      header: 'Node Assignment',
      accessor: (p) => <span className="text-gray-300 font-mono text-xs">{p.nodeName}</span>,
    },
    {
      header: 'netkit Device',
      accessor: (p) => (
        <Badge variant="emerald">{p.netkitInterface || 'netkit0'}</Badge>
      ),
    },
    {
      header: 'Security Identity',
      accessor: (p) => (
        <span className="text-indigo-300 font-mono text-xs">
          Identity #{p.securityIdentity}
        </span>
      ),
    },
    {
      header: 'eBPF Policy',
      accessor: (p) => (
        <div className="flex items-center gap-1.5 text-xs text-emerald-400 font-mono">
          <CheckCircle2 className="w-3.5 h-3.5" />
          <span>Enforced</span>
        </div>
      ),
    },
    {
      header: 'Status',
      accessor: (p) => <StatusDot status={p.status} pulse={p.status === 'Running'} />,
    },
  ]

  const svcColumns: Column<ServiceInfo>[] = [
    {
      header: 'Service Name',
      accessor: (s) => (
        <div className="flex flex-col">
          <span className="text-white font-bold">{s.name}</span>
          <span className="text-[10px] text-gray-500 font-mono">ns: {s.namespace}</span>
        </div>
      ),
    },
    {
      header: 'ClusterIP VIP',
      accessor: (s) => <span className="text-cyan-400 font-mono text-xs">{s.clusterIP}</span>,
    },
    {
      header: 'Ports',
      accessor: (s) => (
        <div className="flex flex-wrap gap-1">
          {s.ports.map((p, idx) => (
            <span
              key={idx}
              className="px-1.5 py-0.5 rounded bg-gray-900 border border-gray-800 text-[10px] text-gray-300 font-mono"
            >
              {p.port}/{p.protocol}
            </span>
          ))}
        </div>
      ),
    },
    {
      header: 'Backends',
      accessor: (s) => (
        <span className="text-gray-300 font-mono text-xs">
          {s.endpointsCount} endpoints
        </span>
      ),
    },
    {
      header: 'eBPF LB Algorithm',
      accessor: (s) => (
        <Badge variant={s.algorithm === 'maglev' ? 'purple' : 'blue'}>
          {s.algorithm}
        </Badge>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      {/* Subtab Navigation */}
      <div className="glass-panel p-2 rounded-xl border border-gray-800 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setSubTab('nodes')}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold font-mono transition-colors ${
              subTab === 'nodes'
                ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40'
                : 'text-gray-400 hover:text-gray-200'
            }`}
          >
            <Server className="w-4 h-4" />
            <span>Nodes ({nodes.length})</span>
          </button>
          <button
            onClick={() => setSubTab('pods')}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold font-mono transition-colors ${
              subTab === 'pods'
                ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40'
                : 'text-gray-400 hover:text-gray-200'
            }`}
          >
            <Boxes className="w-4 h-4" />
            <span>Pods ({pods.length})</span>
          </button>
          <button
            onClick={() => setSubTab('services')}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold font-mono transition-colors ${
              subTab === 'services'
                ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40'
                : 'text-gray-400 hover:text-gray-200'
            }`}
          >
            <Layers className="w-4 h-4" />
            <span>Services ({services.length})</span>
          </button>
        </div>

        <button
          onClick={() => {
            if (subTab === 'nodes') refetchNodes()
            else if (subTab === 'pods') refetchPods()
            else refetchServices()
          }}
          className="text-xs text-cyan-400 hover:text-cyan-300 font-mono px-3"
        >
          {nodesLoading || podsLoading || servicesLoading ? 'Refreshing...' : 'Sync'}
        </button>
      </div>

      {subTab === 'nodes' && (
        <Card
          title="Cluster Nodes"
          subtitle="Host networking, netkit device support, and eBPF attachments"
        >
          <Table<NodeInfo>
            columns={nodeColumns}
            data={nodes}
            keyExtractor={(item) => item.name}
            emptyMessage="No nodes connected"
            id="table-nodes"
          />
        </Card>
      )}

      {subTab === 'pods' && (
        <Card
          title="Pod Workloads"
          subtitle="High-performance direct netkit networking with zero-veth overhead"
        >
          <Table<PodInfo>
            columns={podColumns}
            data={pods}
            keyExtractor={(item) => `${item.namespace}/${item.name}`}
            emptyMessage="No pods found"
            id="table-pods"
          />
        </Card>
      )}

      {subTab === 'services' && (
        <Card
          title="Cluster Services & eBPF Maglev LB"
          subtitle="kube-proxy replacement: in-socket and TCX service VIP translation"
        >
          <Table<ServiceInfo>
            columns={svcColumns}
            data={services}
            keyExtractor={(item) => `${item.namespace}/${item.name}`}
            emptyMessage="No services registered"
            id="table-services"
          />
        </Card>
      )}
    </div>
  )
}
