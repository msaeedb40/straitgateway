import React from 'react'
import { Card } from '../../components/common/Card'
import { Table, Column } from '../../components/common/Table'
import { Badge } from '../../components/common/Badge'
import { StatusDot } from '../../components/common/StatusDot'
import { useRoutes } from '../../hooks/useRoutes'
import { BgpPeerSession, BfdSession, WireguardPeer } from '../../types/networking'
import { KeyRound } from 'lucide-react'

export const BgpPeeringView: React.FC = () => {
  const { bgpPeers, bfdSessions, wireguardPeers, loading, refetch } = useRoutes()

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const bgpColumns: Column<BgpPeerSession>[] = [
    {
      header: 'Peer Address',
      accessor: (p) => <span className="text-white font-bold">{p.peerAddress}</span>,
    },
    {
      header: 'Remote ASN',
      accessor: (p) => <span className="text-cyan-400 font-mono">AS{p.asn}</span>,
    },
    {
      header: 'Session State',
      accessor: (p) => <StatusDot status={p.state} pulse={p.state === 'Established'} />,
    },
    {
      header: 'Uptime',
      accessor: (p) => (
        <span className="text-gray-400 font-mono">
          {Math.floor(p.uptimeSeconds / 60)}m {p.uptimeSeconds % 60}s
        </span>
      ),
    },
    {
      header: 'Prefixes In / Out',
      accessor: (p) => (
        <span className="text-gray-300 font-mono">
          <span className="text-emerald-400">{p.prefixesReceived}</span> rx /{' '}
          <span className="text-cyan-400">{p.prefixesAdvertised}</span> tx
        </span>
      ),
    },
    {
      header: 'Hold / Keepalive',
      accessor: (p) => (
        <span className="text-gray-500 font-mono text-[11px]">
          {p.holdTime}s / {p.keepaliveTime}s
        </span>
      ),
    },
  ]

  const bfdColumns: Column<BfdSession>[] = [
    {
      header: 'Remote Target',
      accessor: (b) => <span className="text-white font-bold">{b.remoteAddress}</span>,
    },
    {
      header: 'BFD State',
      accessor: (b) => <StatusDot status={b.state} pulse={b.state === 'Up'} />,
    },
    {
      header: 'Intervals (Tx / Rx)',
      accessor: (b) => (
        <span className="text-cyan-300 font-mono">
          {b.txIntervalMs}ms / {b.rxIntervalMs}ms
        </span>
      ),
    },
    {
      header: 'Multiplier',
      accessor: (b) => <span className="text-gray-400 font-mono">{b.detectMult}x</span>,
    },
    {
      header: 'Fast Failover',
      accessor: (b) => (
        <Badge variant={b.fastFailoverTriggered ? 'rose' : 'emerald'}>
          {b.fastFailoverTriggered ? 'Triggered (Sub-Second)' : 'Active (Sub-50ms)'}
        </Badge>
      ),
    },
    {
      header: 'Last Transition',
      accessor: (b) => (
        <span className="text-gray-500 text-[11px]">
          {new Date(b.lastTransition).toLocaleTimeString()}
        </span>
      ),
    },
  ]

  const wgColumns: Column<WireguardPeer>[] = [
    {
      header: 'Public Key',
      accessor: (w) => (
        <div className="flex items-center gap-1.5 text-gray-300 font-mono">
          <KeyRound className="w-3.5 h-3.5 text-indigo-400" />
          <span title={w.publicKey}>{w.publicKey.substring(0, 16)}...</span>
        </div>
      ),
    },
    {
      header: 'Endpoint',
      accessor: (w) => <span className="text-white font-mono">{w.endpoint || 'Dynamic'}</span>,
    },
    {
      header: 'Allowed IPs',
      accessor: (w) => (
        <span className="text-cyan-400 font-mono text-[11px]">
          {w.allowedIPs.join(', ')}
        </span>
      ),
    },
    {
      header: 'Latest Handshake',
      accessor: (w) => (
        <span className="text-gray-400 font-mono text-[11px]">
          {w.latestHandshake ? new Date(w.latestHandshake).toLocaleTimeString() : 'Waiting'}
        </span>
      ),
    },
    {
      header: 'Transfer RX / TX',
      accessor: (w) => (
        <span className="text-gray-300 font-mono text-[11px]">
          {formatBytes(w.transferRxBytes)} / {formatBytes(w.transferTxBytes)}
        </span>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      {/* Top BGP / BFD Summary Banner */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card
          title="BGP Dynamic Routing Speakers"
          subtitle="Dynamic prefix exchange with ToR switches and external routers"
        >
          <Table<BgpPeerSession>
            columns={bgpColumns}
            data={bgpPeers}
            keyExtractor={(item) => item.peerAddress}
            emptyMessage="No active BGP peer sessions discovered"
            id="table-bgp-peers"
          />
        </Card>

        <Card
          title="BFD Sub-Second Failure Detection"
          subtitle="Triggers immediate eBPF FIB route invalidation on peer drop"
        >
          <Table<BfdSession>
            columns={bfdColumns}
            data={bfdSessions}
            keyExtractor={(item) => item.remoteAddress}
            emptyMessage="No active BFD sessions configured"
            id="table-bfd-sessions"
          />
        </Card>
      </div>

      {/* WireGuard Mesh Overlay */}
      <Card
        title="WireGuard Inter-Node Overlay Mesh (wg0)"
        subtitle="ChaCha20-Poly1305 encrypted node-to-node and multi-cluster transit tunnels"
        action={
          <button
            onClick={() => refetch()}
            className="text-xs text-cyan-400 hover:text-cyan-300 font-mono"
          >
            {loading ? 'Refreshing...' : 'Sync Peers'}
          </button>
        }
      >
        <Table<WireguardPeer>
          columns={wgColumns}
          data={wireguardPeers}
          keyExtractor={(item) => item.publicKey}
          emptyMessage="No WireGuard peers currently attached to wg0"
          id="table-wireguard-peers"
        />
      </Card>
    </div>
  )
}
