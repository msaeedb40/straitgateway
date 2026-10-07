#include "common.h"
#include "headers.h"
#include "maps.h"
#include "helpers.h"

BPF_MAP_DEF(sg_counters, BPF_MAP_TYPE_PERCPU_ARRAY, __u32, __u64, 64);

SEC("xdp")
int xdp_ingress_func(void *ctx) {
    // Packet counter index 2 = XDP total rx
    __u32 cnt_xdp_idx = 2;
    __u64 *cnt = bpf_map_lookup_elem(&sg_counters, &cnt_xdp_idx);
    if (cnt) {
        __sync_fetch_and_add(cnt, 1);
    }

    return XDP_PASS;
}

char _license[] SEC("license") = "Dual BSD/GPL";
