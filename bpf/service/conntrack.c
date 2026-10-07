#include "common.h"
#include "headers.h"
#include "maps.h"
#include "helpers.h"

BPF_MAP_DEF(sg_conntrack, BPF_MAP_TYPE_LRU_HASH, struct flow_key, struct flow_entry, 262144);
BPF_MAP_DEF(sg_counters, BPF_MAP_TYPE_PERCPU_ARRAY, __u32, __u64, 64);

SEC("tc/conntrack")
int tc_conntrack_func(void *ctx) {
    // Conntrack lookups counter at index 8
    __u32 cnt_ct_idx = 8;
    __u64 *cnt = bpf_map_lookup_elem(&sg_counters, &cnt_ct_idx);
    if (cnt) {
        __sync_fetch_and_add(cnt, 1);
    }
    return TC_ACT_OK;
}

char _license[] SEC("license") = "Dual BSD/GPL";
