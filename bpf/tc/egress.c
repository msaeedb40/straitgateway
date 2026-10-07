#include "common.h"
#include "headers.h"
#include "maps.h"
#include "helpers.h"

BPF_MAP_DEF(sg_counters, BPF_MAP_TYPE_PERCPU_ARRAY, __u32, __u64, 64);

SEC("tc/egress")
int tc_egress_func(void *ctx) {
    // Packet counters index 1 = total tx packets
    __u32 cnt_tx_idx = 1;
    __u64 *cnt_tx = bpf_map_lookup_elem(&sg_counters, &cnt_tx_idx);
    if (cnt_tx) {
        __sync_fetch_and_add(cnt_tx, 1);
    }

    return TC_ACT_OK;
}

char _license[] SEC("license") = "Dual BSD/GPL";
