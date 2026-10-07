import React, { useEffect, useRef, useState } from 'react'
import * as d3 from 'd3'
import { Badge } from '../../components/common/Badge'
import { StatusDot } from '../../components/common/StatusDot'
import { useTopology, TopologyNode, TopologyLink } from '../../hooks/useTopology'
import { Info, Layers } from 'lucide-react'

interface D3Node extends TopologyNode, d3.SimulationNodeDatum {
  x?: number
  y?: number
  fx?: number | null
  fy?: number | null
}

interface D3Link extends d3.SimulationLinkDatum<D3Node> {
  source: D3Node | string
  target: D3Node | string
  type: TopologyLink['type']
  label?: string
}

export const TopologyView: React.FC = () => {
  const { graph, loading, refetch } = useTopology()
  const svgRef = useRef<SVGSVGElement | null>(null)
  const [selectedNode, setSelectedNode] = useState<TopologyNode | null>(null)
  const [filterType, setFilterType] = useState<string>('all')

  useEffect(() => {
    if (!svgRef.current || graph.nodes.length === 0) return

    const width = svgRef.current.clientWidth || 900
    const height = 550

    const svg = d3.select(svgRef.current)
    svg.selectAll('*').remove()

    // Container for zoom/pan
    const g = svg.append('g')

    const zoom = d3
      .zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.2, 3])
      .on('zoom', (event) => {
        g.attr('transform', event.transform)
      })

    svg.call(zoom)

    // Filter nodes if needed
    const nodesData: D3Node[] = graph.nodes
      .filter((n) => (filterType === 'all' ? true : n.type === filterType))
      .map((d) => ({ ...d }))

    const validNodeIds = new Set(nodesData.map((n) => n.id))

    const linksData: D3Link[] = graph.links
      .filter(
        (l) =>
          validNodeIds.has(typeof l.source === 'string' ? l.source : (l.source as D3Node).id) &&
          validNodeIds.has(typeof l.target === 'string' ? l.target : (l.target as D3Node).id)
      )
      .map((d) => ({ ...d }))

    // Simulation
    const simulation = d3
      .forceSimulation<D3Node>(nodesData)
      .force(
        'link',
        d3
          .forceLink<D3Node, D3Link>(linksData)
          .id((d) => d.id)
          .distance((d) => (d.type === 'wireguard' ? 140 : 80))
      )
      .force('charge', d3.forceManyBody().strength(-300))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collision', d3.forceCollide().radius(35))

    // Colors according to node types
    const nodeColors: Record<string, string> = {
      node: '#06b6d4', // Cyan
      pod: '#3b82f6', // Blue
      service: '#6366f1', // Indigo
      gateway: '#10b981', // Emerald
      external: '#f59e0b', // Amber
    }

    // Link styles
    const link = g
      .append('g')
      .selectAll('line')
      .data(linksData)
      .join('line')
      .attr('stroke', (d) =>
        d.type === 'wireguard'
          ? '#6366f1'
          : d.type === 'netkit'
          ? '#06b6d4'
          : '#4b5563'
      )
      .attr('stroke-width', (d) => (d.type === 'wireguard' ? 2 : 1.5))
      .attr('stroke-dasharray', (d) => (d.type === 'wireguard' ? '4 2' : 'none'))
      .attr('stroke-opacity', 0.6)

    // Node groups
    const node = g
      .append('g')
      .selectAll('g')
      .data(nodesData)
      .join('g')
      .style('cursor', 'pointer')
      .on('click', (_, d) => {
        setSelectedNode(d)
      })

    // Drag behavior
    const drag = d3
      .drag<SVGGElement, D3Node>()
      .on('start', (event, d) => {
        if (!event.active) simulation.alphaTarget(0.3).restart()
        d.fx = d.x
        d.fy = d.y
      })
      .on('drag', (event, d) => {
        d.fx = event.x
        d.fy = event.y
      })
      .on('end', (event, d) => {
        if (!event.active) simulation.alphaTarget(0)
        d.fx = null
        d.fy = null
      })

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    node.call(drag as any)

    // Outer circle glow
    node
      .append('circle')
      .attr('r', (d) => (d.type === 'node' ? 22 : d.type === 'gateway' ? 20 : 15))
      .attr('fill', (d) => nodeColors[d.type] || '#9ca3af')
      .attr('fill-opacity', 0.2)
      .attr('stroke', (d) => nodeColors[d.type] || '#9ca3af')
      .attr('stroke-width', 2)

    // Inner core
    node
      .append('circle')
      .attr('r', (d) => (d.type === 'node' ? 8 : 6))
      .attr('fill', (d) => nodeColors[d.type] || '#9ca3af')

    // Node label
    node
      .append('text')
      .text((d) => d.name)
      .attr('x', 0)
      .attr('y', (d) => (d.type === 'node' ? 34 : 26))
      .attr('text-anchor', 'middle')
      .attr('fill', '#e5e7eb')
      .attr('font-size', '11px')
      .attr('font-family', 'monospace')

    simulation.on('tick', () => {
      link
        .attr('x1', (d) => (d.source as D3Node).x || 0)
        .attr('y1', (d) => (d.source as D3Node).y || 0)
        .attr('x2', (d) => (d.target as D3Node).x || 0)
        .attr('y2', (d) => (d.target as D3Node).y || 0)

      node.attr('transform', (d) => `translate(${d.x || 0},${d.y || 0})`)
    })

    return () => {
      simulation.stop()
    }
  }, [graph, filterType])

  return (
    <div className="space-y-4">
      {/* Topology Controls Bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 p-4 glass-panel rounded-xl border border-gray-800">
        <div className="flex items-center gap-2">
          <Layers className="w-5 h-5 text-cyan-400" />
          <h3 className="text-sm font-bold text-white">Dataplane Topology Map</h3>
          <span className="text-xs text-gray-500 font-mono">
            ({graph.nodes.length} nodes, {graph.links.length} links)
          </span>
        </div>

        {/* Filter Badges */}
        <div className="flex items-center gap-1.5 text-xs font-mono">
          {['all', 'node', 'pod', 'service', 'gateway', 'external'].map((t) => (
            <button
              key={t}
              onClick={() => setFilterType(t)}
              className={`px-2.5 py-1 rounded-lg uppercase tracking-wider text-[10px] transition-colors ${
                filterType === t
                  ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40 font-bold'
                  : 'text-gray-400 hover:text-gray-200 bg-gray-900/60'
              }`}
            >
              {t}
            </button>
          ))}
        </div>

        <button
          onClick={() => refetch()}
          className="text-xs text-cyan-400 hover:text-cyan-300 font-mono"
        >
          {loading ? 'Refreshing...' : 'Sync Graph'}
        </button>
      </div>

      {/* Main View: D3 Graph Canvas + Details Sidepane */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-4">
        <div className="lg:col-span-3 glass-panel rounded-xl border border-gray-800 overflow-hidden relative min-h-[550px] bg-gray-950/80">
          <svg ref={svgRef} className="w-full h-[550px]" />

          {/* Canvas Legend */}
          <div className="absolute bottom-4 left-4 p-3 rounded-lg bg-gray-900/80 backdrop-blur-md border border-gray-800 text-[11px] font-mono space-y-1.5">
            <div className="flex items-center gap-2">
              <span className="w-2.5 h-2.5 rounded-full bg-cyan-400" />
              <span className="text-gray-300">Kubernetes Node</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="w-2.5 h-2.5 rounded-full bg-blue-400" />
              <span className="text-gray-300">Pod (netkit device)</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="w-2.5 h-2.5 rounded-full bg-indigo-400" />
              <span className="text-gray-300">Cluster Service</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="w-2.5 h-2.5 rounded-full bg-emerald-400" />
              <span className="text-gray-300">Gateway API Endpoint</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="w-4 h-0.5 border-t border-dashed border-indigo-400" />
              <span className="text-gray-300">WireGuard wg0 Tunnel</span>
            </div>
          </div>
        </div>

        {/* Selected Node Details Pane */}
        <div className="glass-panel rounded-xl border border-gray-800 p-5 flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between pb-3 border-b border-gray-800">
              <h4 className="text-xs font-semibold uppercase tracking-wider text-gray-400 font-mono">
                Component Inspector
              </h4>
              <Info className="w-4 h-4 text-cyan-400" />
            </div>

            {selectedNode ? (
              <div className="mt-4 space-y-3 font-mono text-xs">
                <div>
                  <span className="text-gray-500 uppercase text-[10px]">Name</span>
                  <p className="text-sm font-bold text-white break-all">{selectedNode.name}</p>
                </div>

                <div className="flex items-center justify-between">
                  <span className="text-gray-500 uppercase text-[10px]">Type</span>
                  <Badge variant="cyan">{selectedNode.type}</Badge>
                </div>

                {selectedNode.ip && (
                  <div>
                    <span className="text-gray-500 uppercase text-[10px]">IP Address</span>
                    <p className="text-cyan-300">{selectedNode.ip}</p>
                  </div>
                )}

                <div className="flex items-center justify-between">
                  <span className="text-gray-500 uppercase text-[10px]">Status</span>
                  <StatusDot status={selectedNode.status} />
                </div>

                {selectedNode.details && (
                  <div className="pt-2 border-t border-gray-800 space-y-1.5">
                    <span className="text-gray-500 uppercase text-[10px]">Attributes</span>
                    {Object.entries(selectedNode.details).map(([k, v]) => (
                      <div key={k} className="flex justify-between text-[11px]">
                        <span className="text-gray-400">{k}:</span>
                        <span className="text-gray-200">{String(v)}</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ) : (
              <div className="py-12 text-center text-xs text-gray-500">
                Click any node in the topology map to inspect netkit interfaces, security IDs, and routes.
              </div>
            )}
          </div>

          <div className="pt-4 border-t border-gray-800 text-[10px] text-gray-500 font-mono">
            Interactive drag and zoom enabled.
          </div>
        </div>
      </div>
    </div>
  )
}
