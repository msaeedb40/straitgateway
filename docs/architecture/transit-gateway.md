# Transit Gateway Architecture (tgwd)

## Overview
`tgwd` manages multi-cluster, cross-segment, and inter-network transit topologies. It communicates with `straitd` over a dedicated Unix domain socket (`/run/straitd/api.sock`), ensuring that only `straitd` directly programs the kernel datapath.

## Supported Topologies
1. **Hub-and-Spoke**: Central transit gateway node connects multiple spoke clusters/segments with centralized policy inspection.
2. **Full Mesh**: Every node or gateway peer establishes direct datapath tunnels to every other peer.
3. **Point-to-Point (P2P)**: Direct dedicated link between two specific networks or VPCs.
4. **Hub-to-Hub**: Interconnects multiple hub clusters across regions or cloud providers.
5. **Hybrid**: Combination of hub-and-spoke for branch networks and full mesh for high-throughput core services.
