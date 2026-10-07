#include "common.h"
#include "headers.h"
#include "maps.h"
#include "helpers.h"

SEC("xdp/redirect")
int xdp_redirect_func(void *ctx) {
    return XDP_PASS;
}

char _license[] SEC("license") = "Dual BSD/GPL";
