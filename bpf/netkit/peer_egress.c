#include "common.h"
#include "headers.h"
#include "maps.h"
#include "helpers.h"

BPF_MAP_DEF(sg_counters, BPF_MAP_TYPE_PERCPU_ARRAY, __u32, __u64, 64);

SEC("netkit/peer_egress")
int netkit_peer_egress_func(void *ctx) {
    // Index 4 = pod tx packets
    __u32 cnt_pod_tx = 4;
    __u64 *cnt = bpf_map_lookup_elem(&sg_counters, &cnt_pod_tx);
    if (cnt) {
        __sync_fetch_and_add(cnt, 1);
    }
    return TC_ACT_OK;
}

char _license[] SEC("license") = "Dual BSD/GPL";
