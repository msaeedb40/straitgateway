import React from 'react'
import { Card } from '../../components/common/Card'
import { Table, Column } from '../../components/common/Table'
import { Badge } from '../../components/common/Badge'
import { StatusDot } from '../../components/common/StatusDot'
import { useGateways } from '../../hooks/useGateways'
import { GatewayResource, TransitGatewayResource, HttpRouteResource } from '../../types/gateway'


export const GatewaysView: React.FC = () => {
  const { gateways, transitGateways, httpRoutes, loading, refetch } = useGateways()

  const gwColumns: Column<GatewayResource>[] = [
    {
      header: 'Name',
      accessor: (g) => (
        <div className="flex flex-col">
          <span className="text-white font-bold">{g.name}</span>
          <span className="text-[10px] text-gray-500 font-mono">ns: {g.namespace}</span>
        </div>
      ),
    },
    {
      header: 'GatewayClass',
      accessor: (g) => <span className="text-cyan-400 font-mono">{g.gatewayClassName}</span>,
    },
    {
      header: 'Listeners',
      accessor: (g) => (
        <div className="flex flex-wrap gap-1">
          {g.listeners.map((l, idx) => (
            <span
              key={idx}
              className="px-1.5 py-0.5 rounded bg-gray-900 border border-gray-800 text-[10px] text-gray-300"
            >
              {l.protocol}:{l.port} ({l.routesCount} routes)
            </span>
          ))}
        </div>
      ),
    },
    {
      header: 'Addresses',
      accessor: (g) => (
        <span className="text-gray-300 font-mono text-xs">
          {g.addresses.length > 0 ? g.addresses.join(', ') : 'Pending IP'}
        </span>
      ),
    },
    {
      header: 'Programmed',
      accessor: (g) => (
        <StatusDot status={g.programmed ? 'Ready' : 'Pending'} />
      ),
    },
  ]

  const tgwColumns: Column<TransitGatewayResource>[] = [
    {
      header: 'Transit Gateway',
      accessor: (t) => (
        <div className="flex flex-col">
          <span className="text-white font-bold">{t.name}</span>
          <span className="text-[10px] text-gray-500 font-mono">{t.namespace}</span>
        </div>
      ),
    },
    {
      header: 'Router ID / ASN',
      accessor: (t) => (
        <span className="text-cyan-300 font-mono">
          {t.routerId} (AS{t.asn})
        </span>
      ),
    },
    {
      header: 'VRF Domains',
      accessor: (t) => (
        <div className="flex flex-wrap gap-1">
          {t.vrfs.map((v) => (
            <Badge key={v.id} variant="purple">
              {v.name} (ID: {v.id})
            </Badge>
          ))}
        </div>
      ),
    },
    {
      header: 'Peers & Tunnels',
      accessor: (t) => (
        <span className="text-gray-300 font-mono">
          {t.peersCount} BGP peers • {t.tunnelsCount} WireGuard tunnels
        </span>
      ),
    },
    {
      header: 'Status',
      accessor: (t) => <StatusDot status={t.status} pulse={t.status === 'Active'} />,
    },
  ]

  const routeColumns: Column<HttpRouteResource>[] = [
    {
      header: 'HTTPRoute',
      accessor: (r) => (
        <div className="flex flex-col">
          <span className="text-white font-bold">{r.name}</span>
          <span className="text-[10px] text-gray-500 font-mono">{r.namespace}</span>
        </div>
      ),
    },
    {
      header: 'Hostnames',
      accessor: (r) => (
        <span className="text-cyan-400 font-mono text-xs">
          {r.hostnames.length > 0 ? r.hostnames.join(', ') : '* (all)'}
        </span>
      ),
    },
    {
      header: 'Attached Gateways',
      accessor: (r) => (
        <span className="text-gray-300 font-mono text-xs">
          {r.parentRefs.join(', ')}
        </span>
      ),
    },
    {
      header: 'Rules Count',
      accessor: (r) => <span className="text-gray-400 font-mono">{r.rulesCount}</span>,
    },
  ]

  return (
    <div className="space-y-6">
      {/* Transit Gateway CRDs */}
      <Card
        title="StraitGateway Transit Routers"
        subtitle="Multi-tenant routing domains, VRF isolation, and cross-cluster overlay controllers"
        action={
          <button
            onClick={() => refetch()}
            className="text-xs text-cyan-400 hover:text-cyan-300 font-mono"
          >
            {loading ? 'Refreshing...' : 'Sync Gateways'}
          </button>
        }
      >
        <Table<TransitGatewayResource>
          columns={tgwColumns}
          data={transitGateways}
          keyExtractor={(item) => `${item.namespace}/${item.name}`}
          emptyMessage="No TransitGateway resources registered in the cluster"
          id="table-transit-gateways"
        />
      </Card>

      {/* Gateway API Resources */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card
          title="Gateway API Gateways"
          subtitle="Standard Kubernetes Gateway API v1.6.1 resources"
        >
          <Table<GatewayResource>
            columns={gwColumns}
            data={gateways}
            keyExtractor={(item) => `${item.namespace}/${item.name}`}
            emptyMessage="No Gateway resources configured"
            id="table-gateways"
          />
        </Card>

        <Card
          title="HTTPRoutes & L7 Ingress"
          subtitle="Declarative L7 routing compiled to fast/slow paths"
        >
          <Table<HttpRouteResource>
            columns={routeColumns}
            data={httpRoutes}
            keyExtractor={(item) => `${item.namespace}/${item.name}`}
            emptyMessage="No HTTPRoute resources attached"
            id="table-httproutes"
          />
        </Card>
      </div>
    </div>
  )
}
