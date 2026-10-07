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

SEC("cgroup/sendmsg6")
int sock_sendmsg6_func(struct bpf_sock_addr *ctx) {
    return 1;
}

char _license[] SEC("license") = "Dual BSD/GPL";
