#ifndef __STRAITGATEWAY_COMMON_H__
#define __STRAITGATEWAY_COMMON_H__

typedef unsigned char __u8;
typedef unsigned short __u16;
typedef unsigned int __u32;
typedef unsigned long long __u64;

typedef signed char __s8;
typedef signed short __s16;
typedef signed int __s32;
typedef signed long long __s64;

#ifndef NULL
#define NULL ((void *)0)
#endif

#ifndef SEC
#define SEC(NAME) __attribute__((section(NAME), used))
#endif

#ifndef __always_inline
#define __always_inline inline __attribute__((always_inline))
#endif

#define BPF_MAP_TYPE_HASH 1
#define BPF_MAP_TYPE_ARRAY 2
#define BPF_MAP_TYPE_PROG_ARRAY 3
#define BPF_MAP_TYPE_PERF_EVENT_ARRAY 4
#define BPF_MAP_TYPE_PERCPU_HASH 5
#define BPF_MAP_TYPE_PERCPU_ARRAY 6
#define BPF_MAP_TYPE_LPM_TRIE 11
#define BPF_MAP_TYPE_LRU_HASH 10
#define BPF_MAP_TYPE_RINGBUF 27

#endif /* __STRAITGATEWAY_COMMON_H__ */
