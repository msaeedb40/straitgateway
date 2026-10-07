#include "common.h"
#include "headers.h"
#include "maps.h"
#include "helpers.h"

// Define the maps used in TC ingress
BPF_MAP_DEF(sg_routes, BPF_MAP_TYPE_LPM_TRIE, struct lpm_v4_key, __u32, 65536);
BPF_MAP_DEF(sg_nexthops, BPF_MAP_TYPE_HASH, __u32, struct nexthop_info, 16384);
BPF_MAP_DEF(sg_conntrack, BPF_MAP_TYPE_LRU_HASH, struct flow_key, struct flow_entry, 262144);
BPF_MAP_DEF(sg_policies, BPF_MAP_TYPE_HASH, struct policy_key, __u32, 32768);
BPF_MAP_DEF(sg_counters, BPF_MAP_TYPE_PERCPU_ARRAY, __u32, __u64, 64);

SEC("tc/ingress")
int tc_ingress_func(void *ctx) {
    // Packet counters index 0 = total rx packets
    __u32 cnt_rx_idx = 0;
    __u64 *cnt_rx = bpf_map_lookup_elem(&sg_counters, &cnt_rx_idx);
    if (cnt_rx) {
        __sync_fetch_and_add(cnt_rx, 1);
    }

    // Default action is allow
    return TC_ACT_OK;
}

char _license[] SEC("license") = "Dual BSD/GPL";
