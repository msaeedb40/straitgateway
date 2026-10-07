#ifndef __STRAITGATEWAY_HELPERS_H__
#define __STRAITGATEWAY_HELPERS_H__

#include "common.h"

static void *(*bpf_map_lookup_elem)(void *map, const void *key) = (void *) 1;
static long (*bpf_map_update_elem)(void *map, const void *key, const void *value, __u64 flags) = (void *) 2;
static long (*bpf_map_delete_elem)(void *map, const void *key) = (void *) 3;
static __u64 (*bpf_ktime_get_ns)(void) = (void *) 5;
static long (*bpf_trace_printk)(const char *fmt, __u32 fmt_size, ...) = (void *) 6;
static __u32 (*bpf_get_prandom_u32)(void) = (void *) 7;
static long (*bpf_redirect)(__u32 ifindex, __u64 flags) = (void *) 24;
static long (*bpf_ringbuf_output)(void *ringbuf, void *data, __u64 size, __u64 flags) = (void *) 130;

#define TC_ACT_UNSPEC         (-1)
#define TC_ACT_OK               0
#define TC_ACT_RECLASSIFY       1
#define TC_ACT_SHOT             2
#define TC_ACT_PIPE             3
#define TC_ACT_STOLEN           4
#define TC_ACT_QUEUED           5
#define TC_ACT_REPEAT           6
#define TC_ACT_REDIRECT         7

#define XDP_ABORTED             0
#define XDP_DROP                1
#define XDP_PASS                2
#define XDP_TX                  3
#define XDP_REDIRECT            4

#endif /* __STRAITGATEWAY_HELPERS_H__ */
