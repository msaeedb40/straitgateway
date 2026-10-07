#include "common.h"
#include "headers.h"
#include "maps.h"
#include "helpers.h"

BPF_MAP_DEF(sg_services, BPF_MAP_TYPE_HASH, struct service_key, __u32, 16384);
BPF_MAP_DEF(sg_endpoints, BPF_MAP_TYPE_HASH, __u32, struct service_endpoint, 65536);
BPF_MAP_DEF(sg_counters, BPF_MAP_TYPE_PERCPU_ARRAY, __u32, __u64, 64);

SEC("tc/service_lb")
int tc_service_lb_func(void *ctx) {
    // Service LB metric counter at index 6
    __u32 cnt_lb_idx = 6;
    __u64 *cnt = bpf_map_lookup_elem(&sg_counters, &cnt_lb_idx);
    if (cnt) {
        __sync_fetch_and_add(cnt, 1);
    }
    return TC_ACT_OK;
}

char _license[] SEC("license") = "Dual BSD/GPL";
