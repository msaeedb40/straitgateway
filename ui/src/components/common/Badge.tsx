import React from 'react'

export interface BadgeProps {
  children: React.ReactNode
  variant?: 'cyan' | 'blue' | 'indigo' | 'emerald' | 'amber' | 'rose' | 'purple' | 'gray'
  size?: 'sm' | 'md'
  id?: string
}

const variantStyles: Record<string, string> = {
  cyan: 'bg-cyan-500/10 text-cyan-400 border-cyan-500/30',
  blue: 'bg-blue-500/10 text-blue-400 border-blue-500/30',
  indigo: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/30',
  emerald: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30',
  amber: 'bg-amber-500/10 text-amber-400 border-amber-500/30',
  rose: 'bg-rose-500/10 text-rose-400 border-rose-500/30',
  purple: 'bg-purple-500/10 text-purple-400 border-purple-500/30',
  gray: 'bg-gray-500/10 text-gray-400 border-gray-500/30',
}

export const Badge: React.FC<BadgeProps> = ({ children, variant = 'gray', size = 'sm', id }) => {
  const sizeClasses = size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-2.5 py-1 text-sm'
  const colorClass = variantStyles[variant] || variantStyles.gray

  return (
    <span
      id={id}
      className={`inline-flex items-center font-medium rounded-full border ${colorClass} ${sizeClasses}`}
    >
      {children}
    </span>
  )
}
