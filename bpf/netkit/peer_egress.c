#include "common.h"
#include "headers.h"
#include "maps.h"
#include "helpers.h"

BPF_MAP_DEF(sg_policies, BPF_MAP_TYPE_HASH, struct policy_key, __u32, 32768);
BPF_MAP_DEF(sg_identities, BPF_MAP_TYPE_HASH, __u32, __u32, 65536);
BPF_MAP_DEF(sg_counters, BPF_MAP_TYPE_PERCPU_ARRAY, __u32, __u64, 64);

SEC("netkit/peer_egress")
int netkit_peer_egress_func(struct __sk_buff *skb) {
    // Index 4 = pod tx packets
    __u32 cnt_pod_tx = 4;
    __u64 *cnt = bpf_map_lookup_elem(&sg_counters, &cnt_pod_tx);
    if (cnt) {
        __sync_fetch_and_add(cnt, 1);
    }

    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end) {
        return TC_ACT_OK;
    }

    if (eth->h_proto != __builtin_bswap16(ETH_P_IP)) {
        return TC_ACT_OK;
    }

    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end) {
        return TC_ACT_OK;
    }

    __u32 *src_id = bpf_map_lookup_elem(&sg_identities, &iph->saddr);
    __u32 *dst_id = bpf_map_lookup_elem(&sg_identities, &iph->daddr);

    if (src_id) {
        __u32 effective_dst_id = dst_id ? *dst_id : 2; // IdentityWorld = 2
        __u16 dst_port = 0;
        void *l4 = (void *)(iph + 1);

        if (iph->protocol == SG_PROTO_TCP) {
            struct tcphdr *th = l4;
            if ((void *)(th + 1) <= data_end) {
                dst_port = __builtin_bswap16(th->dest);
            }
        } else if (iph->protocol == SG_PROTO_UDP) {
            struct udphdr *uh = l4;
            if ((void *)(uh + 1) <= data_end) {
                dst_port = __builtin_bswap16(uh->dest);
            }
        }

        struct policy_key pkey = {
            .src_id = *src_id,
            .dst_id = effective_dst_id,
            .dst_port = dst_port,
            .proto = iph->protocol,
            .pad = 0,
        };
        __u32 *action = bpf_map_lookup_elem(&sg_policies, &pkey);

        if (!action) {
            pkey.dst_port = 0;
            action = bpf_map_lookup_elem(&sg_policies, &pkey);
        }
        if (!action) {
            pkey.proto = 0;
            action = bpf_map_lookup_elem(&sg_policies, &pkey);
        }

        if (action && *action == POLICY_DROP) {
            __u32 cnt_drop_idx = 5;
            __u64 *cnt_drop = bpf_map_lookup_elem(&sg_counters, &cnt_drop_idx);
            if (cnt_drop) {
                __sync_fetch_and_add(cnt_drop, 1);
            }
            return TC_ACT_SHOT;
        }
    }

    return TC_ACT_OK;
}

char _license[] SEC("license") = "Dual BSD/GPL";
