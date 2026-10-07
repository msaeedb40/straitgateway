import React, { useState } from 'react'
import { X, Save, AlertCircle } from 'lucide-react'
import { getRuntimeConfig, updateRuntimeConfig } from '../../config/runtime'

interface SettingsModalProps {
  isOpen: boolean
  onClose: () => void
  onSave: () => void
}

export const SettingsModal: React.FC<SettingsModalProps> = ({ isOpen, onClose, onSave }) => {
  const current = getRuntimeConfig()
  const [apiUrl, setApiUrl] = useState(current.apiBaseUrl)
  const [refreshInterval, setRefreshInterval] = useState(current.refreshIntervalMs)

  if (!isOpen) return null

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault()
    updateRuntimeConfig({
      apiBaseUrl: apiUrl,
      refreshIntervalMs: refreshInterval,
    })
    onSave()
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm">
      <div className="w-full max-w-md bg-gray-900 border border-gray-800 rounded-xl shadow-2xl p-6 relative">
        <div className="flex items-center justify-between pb-4 border-b border-gray-800">
          <h3 className="text-base font-bold text-white">Runtime Parameter Configuration</h3>
          <button
            onClick={onClose}
            className="p-1 rounded-lg text-gray-400 hover:text-white hover:bg-gray-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSave} className="mt-4 space-y-4">
          <div>
            <label className="block text-xs font-medium text-gray-300 uppercase tracking-wider mb-1">
              StraitGateway API Base URL
            </label>
            <input
              type="text"
              id="input-api-url"
              value={apiUrl}
              onChange={(e) => setApiUrl(e.target.value)}
              placeholder="/api/v1"
              className="w-full px-3 py-2 bg-gray-950 border border-gray-800 rounded-lg text-sm text-gray-100 font-mono focus:outline-none focus:border-cyan-500"
            />
            <p className="text-[11px] text-gray-500 mt-1">
              Authoritative endpoint for Kubernetes & dataplane state (no mock data).
            </p>
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-300 uppercase tracking-wider mb-1">
              Telemetry Polling Interval (ms)
            </label>
            <input
              type="number"
              id="input-refresh-interval"
              min={1000}
              max={60000}
              step={1000}
              value={refreshInterval}
              onChange={(e) => setRefreshInterval(Number(e.target.value))}
              className="w-full px-3 py-2 bg-gray-950 border border-gray-800 rounded-lg text-sm text-gray-100 font-mono focus:outline-none focus:border-cyan-500"
            />
          </div>

          <div className="p-3 bg-cyan-950/30 border border-cyan-800/40 rounded-lg flex items-start gap-2.5 text-xs text-cyan-300">
            <AlertCircle className="w-4 h-4 shrink-0 mt-0.5 text-cyan-400" />
            <p>
              Settings are stored in browser local storage and take effect across all custom hooks immediately.
            </p>
          </div>

          <div className="flex justify-end gap-3 pt-3 border-t border-gray-800">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-xs font-medium text-gray-400 hover:text-white hover:bg-gray-800 rounded-lg transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              id="btn-save-settings"
              className="flex items-center gap-2 px-4 py-2 bg-cyan-500 hover:bg-cyan-400 text-gray-950 text-xs font-bold rounded-lg transition-colors"
            >
              <Save className="w-4 h-4" />
              Save Configuration
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
