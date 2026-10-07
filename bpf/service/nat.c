#include "common.h"
#include "headers.h"
#include "maps.h"
#include "helpers.h"

BPF_MAP_DEF(sg_counters, BPF_MAP_TYPE_PERCPU_ARRAY, __u32, __u64, 64);

SEC("tc/nat")
int tc_nat_func(void *ctx) {
    // NAT translation counter at index 7
    __u32 cnt_nat_idx = 7;
    __u64 *cnt = bpf_map_lookup_elem(&sg_counters, &cnt_nat_idx);
    if (cnt) {
        __sync_fetch_and_add(cnt, 1);
    }
    return TC_ACT_OK;
}

char _license[] SEC("license") = "Dual BSD/GPL";
