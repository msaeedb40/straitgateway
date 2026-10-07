import React from 'react'
import {
  LayoutDashboard,
  Network,
  Activity,
  Boxes,
  ShieldCheck,
  Radio,
  Sliders,
  Server,
} from 'lucide-react'

export type NavTab =
  | 'overview'
  | 'topology'
  | 'flows'
  | 'routes'
  | 'routing-bgp'
  | 'gateways'
  | 'security'
  | 'ebpf'
  | 'k8s-workloads'

interface SidebarProps {
  activeTab: NavTab
  onSelectTab: (tab: NavTab) => void
  clusterName: string
}

export const Sidebar: React.FC<SidebarProps> = ({ activeTab, onSelectTab, clusterName }) => {
  const navSections = [
    {
      group: 'Insights & Health',
      items: [
        { id: 'overview' as NavTab, label: 'Overview', icon: LayoutDashboard },
        { id: 'topology' as NavTab, label: 'Topology Mesh', icon: Network },
        { id: 'flows' as NavTab, label: 'Traffic & Flows', icon: Activity },
      ],
    },
    {
      group: 'Transit & Networking',
      items: [
        { id: 'gateways' as NavTab, label: 'Gateways & Routes', icon: Sliders },
        { id: 'routing-bgp' as NavTab, label: 'BGP & BFD Peering', icon: Radio },
        { id: 'routes' as NavTab, label: 'Kernel FIB & Routes', icon: Server },
      ],
    },
    {
      group: 'Security & Dataplane',
      items: [
        { id: 'security' as NavTab, label: 'Security Policies', icon: ShieldCheck },
        { id: 'ebpf' as NavTab, label: 'eBPF Programs', icon: Activity },
        { id: 'k8s-workloads' as NavTab, label: 'K8s Workloads', icon: Boxes },
      ],
    },
  ]

  return (
    <aside className="w-64 bg-gray-950 border-r border-gray-800/80 flex flex-col h-screen shrink-0">
      {/* Brand */}
      <div className="h-16 flex items-center px-6 gap-3 border-b border-gray-800/80 bg-gray-900/30">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-tr from-cyan-500 to-indigo-600 flex items-center justify-center shadow-lg shadow-cyan-500/20">
          <Network className="w-5 h-5 text-white" />
        </div>
        <div>
          <h1 className="text-sm font-bold tracking-tight text-white flex items-center gap-1.5">
            StraitGateway
            <span className="text-[10px] font-mono px-1.5 py-0.2 bg-cyan-500/20 text-cyan-300 rounded border border-cyan-500/30">
              v1.0
            </span>
          </h1>
          <p className="text-[11px] text-gray-400 font-mono truncate max-w-[130px]">{clusterName}</p>
        </div>
      </div>

      {/* Nav List */}
      <nav className="flex-1 overflow-y-auto px-3 py-4 space-y-6">
        {navSections.map((sec, idx) => (
          <div key={idx}>
            <div className="px-3 mb-2 text-[10px] font-semibold text-gray-400 uppercase tracking-wider font-mono">
              {sec.group}
            </div>
            <div className="space-y-1">
              {sec.items.map((item) => {
                const Icon = item.icon
                const isActive = activeTab === item.id
                return (
                  <button
                    key={item.id}
                    id={`nav-btn-${item.id}`}
                    onClick={() => onSelectTab(item.id)}
                    className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-medium transition-all ${
                      isActive
                        ? 'bg-gradient-to-r from-cyan-500/20 to-indigo-500/20 text-cyan-300 border border-cyan-500/30 font-semibold shadow-sm'
                        : 'text-gray-400 hover:text-gray-200 hover:bg-gray-900/60'
                    }`}
                  >
                    <Icon className={`w-4 h-4 ${isActive ? 'text-cyan-400' : 'text-gray-400'}`} />
                    <span>{item.label}</span>
                  </button>
                )
              })}
            </div>
          </div>
        ))}
      </nav>

      {/* Footer Info */}
      <div className="p-4 border-t border-gray-800/80 bg-gray-950/60 text-[11px] text-gray-400 font-mono space-y-1">
        <div className="flex justify-between">
          <span>Dataplane:</span>
          <span className="text-gray-300">eBPF + netkit</span>
        </div>
        <div className="flex justify-between">
          <span>Kernel:</span>
          <span className="text-gray-300">Linux 6.12+ LTS</span>
        </div>
      </div>
    </aside>
  )
}
