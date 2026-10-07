import React from 'react'

export interface Column<T> {
  header: string
  accessor: (item: T) => React.ReactNode
  className?: string
}

export interface TableProps<T> {
  columns: Column<T>[]
  data: T[]
  emptyMessage?: string
  keyExtractor: (item: T, index: number) => string
  id?: string
}

export function Table<T>({
  columns,
  data,
  emptyMessage = 'No data available',
  keyExtractor,
  id,
}: TableProps<T>): React.ReactElement {
  return (
    <div id={id} className="overflow-x-auto w-full">
      <table className="w-full text-left text-sm text-gray-300">
        <thead className="text-xs uppercase bg-gray-900/60 text-gray-400 border-b border-gray-800">
          <tr>
            {columns.map((col, idx) => (
              <th key={idx} scope="col" className={`px-4 py-3 font-semibold ${col.className || ''}`}>
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-800/60 font-mono text-xs">
          {data.length === 0 ? (
            <tr>
              <td colSpan={columns.length} className="px-4 py-8 text-center text-gray-500 font-sans">
                {emptyMessage}
              </td>
            </tr>
          ) : (
            data.map((item, rowIdx) => (
              <tr key={keyExtractor(item, rowIdx)} className="hover:bg-gray-800/40 transition-colors">
                {columns.map((col, colIdx) => (
                  <td key={colIdx} className={`px-4 py-3 whitespace-nowrap ${col.className || ''}`}>
                    {col.accessor(item)}
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  )
}
