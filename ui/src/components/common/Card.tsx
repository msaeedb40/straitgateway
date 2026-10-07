import React from 'react'

export interface CardProps {
  children: React.ReactNode
  title?: string
  subtitle?: string
  action?: React.ReactNode
  className?: string
  id?: string
}

export const Card: React.FC<CardProps> = ({ children, title, subtitle, action, className = '', id }) => {
  return (
    <div
      id={id}
      className={`glass-panel rounded-xl p-5 border border-gray-800/80 shadow-lg ${className}`}
    >
      {(title || action) && (
        <div className="flex items-center justify-between pb-4 mb-4 border-b border-gray-800/60">
          <div>
            {title && <h3 className="text-base font-semibold text-gray-100">{title}</h3>}
            {subtitle && <p className="text-xs text-gray-400 mt-0.5">{subtitle}</p>}
          </div>
          {action && <div>{action}</div>}
        </div>
      )}
      {children}
    </div>
  )
}
