#include "common.h"
#include "headers.h"
#include "maps.h"
#include "helpers.h"

struct bpf_sock_addr {
    __u32 user_family;
    __u32 user_ip4;
    __u32 user_ip6[4];
    __u32 user_port;
    __u32 family;
    __u32 type;
    __u32 protocol;
    __u32 msg_src_ip4;
    __u32 msg_src_ip6[4];
    void *sk;
};

BPF_MAP_DEF(sg_services, BPF_MAP_TYPE_HASH, struct service_key, __u32, 16384);
BPF_MAP_DEF(sg_endpoints, BPF_MAP_TYPE_HASH, __u32, struct service_endpoint, 65536);

SEC("cgroup/sendmsg4")
int sock_sendmsg4_func(struct bpf_sock_addr *ctx) {
    if (!ctx)
        return 1;

    struct service_key key = {
        .ip = ctx->user_ip4,
        .port = (__u16)ctx->user_port,
        .proto = (__u8)ctx->protocol,
        .pad = 0,
    };

    __u32 *ep_id = bpf_map_lookup_elem(&sg_services, &key);
    if (!ep_id) {
        return 1;
    }

    struct service_endpoint *ep = bpf_map_lookup_elem(&sg_endpoints, ep_id);
    if (!ep || ep->state != SG_EP_ACTIVE) {
        return 1;
    }

    ctx->user_ip4 = ep->ip;
    ctx->user_port = (__u32)ep->port;
    return 1;
}

char _license[] SEC("license") = "Dual BSD/GPL";
