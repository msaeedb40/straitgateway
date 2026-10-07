import React from 'react'
import { RefreshCw, Settings, Wifi, WifiOff } from 'lucide-react'
import { StatusDot } from '../common/StatusDot'

interface HeaderProps {
  title: string
  clusterName: string
  environment: string
  refreshIntervalMs: number
  onUpdateRefreshInterval: (ms: number) => void
  onManualRefresh: () => void
  isRefreshing: boolean
  isOnline: boolean
  lastUpdated: Date | null
  onOpenSettings: () => void
}

export const Header: React.FC<HeaderProps> = ({
  title,
  clusterName,
  environment,
  refreshIntervalMs,
  onUpdateRefreshInterval,
  onManualRefresh,
  isRefreshing,
  isOnline,
  lastUpdated,
  onOpenSettings,
}) => {
  return (
    <header className="h-16 px-6 border-b border-gray-800/80 bg-gray-950/70 backdrop-blur-md flex items-center justify-between shrink-0">
      <div className="flex items-center gap-3">
        <h2 className="text-lg font-bold text-white tracking-tight">{title}</h2>
        <span className="text-xs font-mono px-2 py-0.5 rounded-full bg-gray-800 text-gray-300 border border-gray-700/60">
          {environment}
        </span>
      </div>

      <div className="flex items-center gap-4">
        {/* Status Indicator */}
        <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-gray-900/80 border border-gray-800 text-xs">
          {isOnline ? (
            <>
              <Wifi className="w-3.5 h-3.5 text-emerald-400" />
              <StatusDot status="Active" pulse />
              <span className="text-gray-400 font-mono text-[11px] hidden sm:inline">
                {clusterName}
              </span>
            </>
          ) : (
            <>
              <WifiOff className="w-3.5 h-3.5 text-rose-400" />
              <StatusDot status="Offline" />
              <span className="text-rose-400 font-mono text-[11px]">Reconnecting</span>
            </>
          )}
        </div>

        {/* Polling Interval Selector */}
        <div className="flex items-center gap-1.5 bg-gray-900/80 border border-gray-800 rounded-lg p-1 text-xs">
          <span className="text-[11px] text-gray-400 px-2 font-mono">Sync:</span>
          {[2000, 5000, 10000].map((ms) => (
            <button
              key={ms}
              id={`refresh-interval-${ms}`}
              onClick={() => onUpdateRefreshInterval(ms)}
              className={`px-2 py-0.5 rounded text-[11px] font-mono transition-colors ${
                refreshIntervalMs === ms
                  ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40 font-bold'
                  : 'text-gray-400 hover:text-gray-200'
              }`}
            >
              {ms / 1000}s
            </button>
          ))}
        </div>

        {/* Manual Refresh Button */}
        <button
          id="btn-manual-refresh"
          onClick={onManualRefresh}
          disabled={isRefreshing}
          title={lastUpdated ? `Last updated: ${lastUpdated.toLocaleTimeString()}` : 'Refresh'}
          className="p-2 rounded-lg bg-gray-900 hover:bg-gray-800 text-gray-300 border border-gray-800 transition-all active:scale-95 disabled:opacity-50"
        >
          <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin text-cyan-400' : ''}`} />
        </button>

        {/* Settings Button */}
        <button
          id="btn-open-settings"
          onClick={onOpenSettings}
          title="Runtime Settings"
          className="p-2 rounded-lg bg-gray-900 hover:bg-gray-800 text-gray-300 border border-gray-800 transition-all hover:text-white"
        >
          <Settings className="w-4 h-4" />
        </button>
      </div>
    </header>
  )
}
