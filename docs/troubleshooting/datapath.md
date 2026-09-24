# StraitGateway Datapath Troubleshooting & Diagnostics Guide

## Command Line Diagnostics (`sgctl`)

### 1. Check Agent & Datapath Status
```bash
# Verify straitd agent health and socket connectivity
sgctl status

# Detailed component health
sgctl health
```

### 2. Inspect Active Network Flows
```bash
# View live 5-tuple flow records from kernel ring buffer
sgctl flow --tail 50
```

### 3. Live Packet Capture (`sgpktcap`)
```bash
# Capture packets on a specific pod interface
sgpktcap capture --interface nk-pod1 --filter "tcp and port 80"

# Follow capture in real-time
sgpktcap follow

# Export capture to pcap format for Wireshark inspection
sgpktcap export --output /tmp/capture.pcap
```

### 4. Low-Level eBPF Inspection (`bpftool`)
```bash
# Inspect loaded BPF programs
bpftool prog show

# Dump service VIP translation table
bpftool map dump pinned /sys/fs/bpf/sg_services

# Dump policy decision map
bpftool map dump pinned /sys/fs/bpf/sg_policy
```
