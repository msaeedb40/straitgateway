import React, { useState } from 'react'
import { Sidebar, NavTab } from './components/navigation/Sidebar'
import { Header } from './components/navigation/Header'
import { SettingsModal } from './components/common/SettingsModal'
import { OverviewView } from './features/insights/OverviewView'
import { TopologyView } from './features/topology/TopologyView'
import { FlowsView } from './features/insights/FlowsView'
import { RoutesView } from './features/networking/RoutesView'
import { BgpPeeringView } from './features/networking/BgpPeeringView'
import { GatewaysView } from './features/gateways/GatewaysView'
import { SecurityPoliciesView } from './features/security/SecurityPoliciesView'
import { EbpfProgramsView } from './features/networking/EbpfProgramsView'
import { WorkloadsView } from './features/cluster/WorkloadsView'
import { useCluster } from './hooks/useCluster'
import { getRuntimeConfig, updateRuntimeConfig } from './config/runtime'

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<NavTab>('overview')
  const [isSettingsOpen, setIsSettingsOpen] = useState(false)
  const config = getRuntimeConfig()
  const [refreshInterval, setRefreshInterval] = useState(config.refreshIntervalMs)

  const { loading, error, lastUpdated, refetch } = useCluster(refreshInterval)

  const tabTitles: Record<NavTab, string> = {
    overview: 'Cluster Insights & Dataplane Health',
    topology: 'Real-time Dataplane Topology & Mesh',
    flows: 'Kernel Flow Interceptor & Telemetry',
    routes: 'Forwarding Information Base (FIB)',
    'routing-bgp': 'BGP Dynamic Routing & BFD Heartbeats',
    gateways: 'Kubernetes Gateway API & Transit Gateways',
    security: 'Zero-Trust Security & Identity Policies',
    ebpf: 'eBPF Program Verifier & Hook Attachments',
    'k8s-workloads': 'Kubernetes Workloads & netkit Pods',
  }

  const handleUpdateRefreshInterval = (ms: number) => {
    setRefreshInterval(ms)
    updateRuntimeConfig({ refreshIntervalMs: ms })
  }

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-[#0b0f17] text-gray-100 font-sans antialiased">
      {/* Sidebar */}
      <Sidebar
        activeTab={activeTab}
        onSelectTab={setActiveTab}
        clusterName={config.clusterName}
      />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col h-full overflow-hidden">
        {/* Header */}
        <Header
          title={tabTitles[activeTab]}
          clusterName={config.clusterName}
          environment={config.environment}
          refreshIntervalMs={refreshInterval}
          onUpdateRefreshInterval={handleUpdateRefreshInterval}
          onManualRefresh={refetch}
          isRefreshing={loading}
          isOnline={!error}
          lastUpdated={lastUpdated}
          onOpenSettings={() => setIsSettingsOpen(true)}
        />

        {/* Tab Content */}
        <main className="flex-1 overflow-y-auto p-6 bg-[#090d14]/60">
          <div className="max-w-7xl mx-auto">
            {activeTab === 'overview' && (
              <OverviewView onNavigate={(tab) => setActiveTab(tab as NavTab)} />
            )}
            {activeTab === 'topology' && <TopologyView />}
            {activeTab === 'flows' && <FlowsView />}
            {activeTab === 'routes' && <RoutesView />}
            {activeTab === 'routing-bgp' && <BgpPeeringView />}
            {activeTab === 'gateways' && <GatewaysView />}
            {activeTab === 'security' && <SecurityPoliciesView />}
            {activeTab === 'ebpf' && <EbpfProgramsView />}
            {activeTab === 'k8s-workloads' && <WorkloadsView />}
          </div>
        </main>
      </div>

      {/* Runtime Settings Modal */}
      <SettingsModal
        isOpen={isSettingsOpen}
        onClose={() => setIsSettingsOpen(false)}
        onSave={() => {
          setRefreshInterval(getRuntimeConfig().refreshIntervalMs)
          refetch()
        }}
      />
    </div>
  )
}

export default App
