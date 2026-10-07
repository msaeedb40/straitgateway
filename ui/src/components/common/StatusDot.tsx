import React from 'react'

export interface StatusDotProps {
  status: 'Ready' | 'Healthy' | 'Active' | 'Established' | 'Up' | 'Running' | 'Degraded' | 'Pending' | 'Down' | 'Failed' | 'Critical' | string
  pulse?: boolean
  id?: string
}

export const StatusDot: React.FC<StatusDotProps> = ({ status, pulse = false, id }) => {
  const isGood = ['Ready', 'Healthy', 'Active', 'Established', 'Up', 'Running'].includes(status)
  const isWarn = ['Degraded', 'Pending', 'OpenConfirm', 'Connect'].includes(status)

  const colorBg = isGood ? 'bg-emerald-400' : isWarn ? 'bg-amber-400' : 'bg-rose-400'
  const colorGlow = isGood ? 'shadow-emerald-500/50' : isWarn ? 'shadow-amber-500/50' : 'shadow-rose-500/50'

  return (
    <span id={id} className="inline-flex items-center gap-1.5">
      <span className="relative flex h-2.5 w-2.5">
        {pulse && isGood && (
          <span className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${colorBg}`} />
        )}
        <span className={`relative inline-flex rounded-full h-2.5 w-2.5 shadow-sm ${colorBg} ${colorGlow}`} />
      </span>
      <span className="text-xs text-gray-300 font-medium">{status}</span>
    </span>
  )
}
