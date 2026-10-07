import React from 'react'
import { Card } from './Card'

export interface MetricCardProps {
  label: string
  value: string | number
  unit?: string
  change?: string
  trend?: 'up' | 'down' | 'neutral'
  icon?: React.ReactNode
  variant?: 'cyan' | 'blue' | 'indigo' | 'emerald' | 'rose'
  id?: string
}

export const MetricCard: React.FC<MetricCardProps> = ({
  label,
  value,
  unit,
  change,
  trend = 'neutral',
  icon,
  variant = 'cyan',
  id,
}) => {
  const accentBorders: Record<string, string> = {
    cyan: 'border-l-4 border-l-cyan-500',
    blue: 'border-l-4 border-l-blue-500',
    indigo: 'border-l-4 border-l-indigo-500',
    emerald: 'border-l-4 border-l-emerald-500',
    rose: 'border-l-4 border-l-rose-500',
  }

  const borderClass = accentBorders[variant] || accentBorders.cyan

  return (
    <Card id={id} className={`glass-panel-hover ${borderClass}`}>
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs font-medium text-gray-400 uppercase tracking-wider">{label}</p>
          <div className="flex items-baseline gap-1 mt-2">
            <span className="text-2xl font-bold tracking-tight text-white">{value}</span>
            {unit && <span className="text-xs font-semibold text-gray-400">{unit}</span>}
          </div>
          {change && (
            <p
              className={`text-xs mt-1.5 font-medium ${
                trend === 'up'
                  ? 'text-emerald-400'
                  : trend === 'down'
                  ? 'text-rose-400'
                  : 'text-gray-400'
              }`}
            >
              {change}
            </p>
          )}
        </div>
        {icon && (
          <div className="p-2.5 rounded-lg bg-gray-800/60 text-gray-300 border border-gray-700/50">
            {icon}
          </div>
        )}
      </div>
    </Card>
  )
}
