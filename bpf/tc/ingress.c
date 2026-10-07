#include "common.h"
#include "headers.h"
#include "maps.h"
#include "helpers.h"

// Define the maps used in TC ingress
BPF_MAP_DEF(sg_routes, BPF_MAP_TYPE_LPM_TRIE, struct lpm_v4_key, __u32, 65536);
BPF_MAP_DEF(sg_nexthops, BPF_MAP_TYPE_HASH, __u32, struct nexthop_info, 16384);
BPF_MAP_DEF(sg_conntrack, BPF_MAP_TYPE_LRU_HASH, struct flow_key, struct flow_entry, 262144);
BPF_MAP_DEF(sg_policies, BPF_MAP_TYPE_HASH, struct policy_key, __u32, 32768);
BPF_MAP_DEF(sg_identities, BPF_MAP_TYPE_HASH, __u32, __u32, 65536);
BPF_MAP_DEF(sg_counters, BPF_MAP_TYPE_PERCPU_ARRAY, __u32, __u64, 64);

SEC("tc/ingress")
int tc_ingress_func(struct __sk_buff *skb) {
    // Packet counters index 0 = total rx packets
    __u32 cnt_rx_idx = 0;
    __u64 *cnt_rx = bpf_map_lookup_elem(&sg_counters, &cnt_rx_idx);
    if (cnt_rx) {
        __sync_fetch_and_add(cnt_rx, 1);
    }

    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    // Bounds check Ethernet header
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end) {
        return TC_ACT_OK;
    }

    // Only process IPv4 for transit forwarding & security policy
    if (eth->h_proto != __builtin_bswap16(ETH_P_IP)) {
        return TC_ACT_OK;
    }

    // Bounds check IPv4 header
    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end) {
        return TC_ACT_OK;
    }

    // 1. Identity & Security Policy Evaluation
    __u32 *src_id = bpf_map_lookup_elem(&sg_identities, &iph->saddr);
    __u32 *dst_id = bpf_map_lookup_elem(&sg_identities, &iph->daddr);

    // If destination has a registered workload identity, evaluate policies
    if (dst_id) {
        __u32 effective_src_id = src_id ? *src_id : 2; // IdentityWorld = 2
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

        // Exact match lookup: {src, dst, dst_port, proto}
        struct policy_key pkey = {
            .src_id = effective_src_id,
            .dst_id = *dst_id,
            .dst_port = dst_port,
            .proto = iph->protocol,
            .pad = 0,
        };
        __u32 *action = bpf_map_lookup_elem(&sg_policies, &pkey);

        // Wildcard port lookup: {src, dst, 0, proto}
        if (!action) {
            pkey.dst_port = 0;
            action = bpf_map_lookup_elem(&sg_policies, &pkey);
        }

        // Wildcard proto lookup: {src, dst, 0, 0}
        if (!action) {
            pkey.proto = 0;
            action = bpf_map_lookup_elem(&sg_policies, &pkey);
        }

        if (action && *action == POLICY_DROP) {
            // Drop counter at index 5 (security policy drops)
            __u32 cnt_drop_idx = 5;
            __u64 *cnt_drop = bpf_map_lookup_elem(&sg_counters, &cnt_drop_idx);
            if (cnt_drop) {
                __sync_fetch_and_add(cnt_drop, 1);
            }
            return TC_ACT_SHOT;
        }
    }

    // 2. LPM Route Lookup
    struct lpm_v4_key rkey;
    rkey.prefixlen = 32;
    rkey.addr = iph->daddr;


    __u32 *nh_id = bpf_map_lookup_elem(&sg_routes, &rkey);
    if (!nh_id) {
        // Fallback / default pass
        return TC_ACT_OK;
    }

    // Next-hop resolution
    struct nexthop_info *nh = bpf_map_lookup_elem(&sg_nexthops, nh_id);
    if (!nh) {
        return TC_ACT_OK;
    }

    // Check if next hop has target interface for forwarding
    if (nh->ifindex > 0) {
        // Decrement TTL
        if (iph->ttl <= 1) {
            // Drop expired packet
            return TC_ACT_SHOT;
        }
        iph->ttl--;

        // Rewrite destination MAC if next-hop provides it
        if (nh->mac[0] != 0 || nh->mac[1] != 0 || nh->mac[2] != 0 ||
            nh->mac[3] != 0 || nh->mac[4] != 0 || nh->mac[5] != 0) {
            #pragma unroll
            for (int i = 0; i < 6; i++) {
                eth->h_dest[i] = nh->mac[i];
            }
        }

        // Increment routed counter at index 3
        __u32 cnt_routed_idx = 3;
        __u64 *cnt_routed = bpf_map_lookup_elem(&sg_counters, &cnt_routed_idx);
        if (cnt_routed) {
            __sync_fetch_and_add(cnt_routed, 1);
        }

        // Direct packet redirection to destination interface (veth, netkit, or wg0)
        return bpf_redirect(nh->ifindex, 0);
    }

    return TC_ACT_OK;
}

char _license[] SEC("license") = "Dual BSD/GPL";
