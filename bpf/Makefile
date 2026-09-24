CC      := $(shell which clang-22 clang-21 clang 2>/dev/null | head -n1)
LLC     := $(shell which llc-22 llc-21 llc 2>/dev/null | head -n1)
BPFTOOL ?= bpftool

ifeq ($(CC),)
$(error No suitable clang found. Please install clang-22, clang-21, or clang.)
endif

ARCH ?= $(shell uname -m | sed 's/x86_64/x86/' | sed 's/aarch64/arm64/')

SRC_DIR := .
OUT_DIR := out

# BPF compiler flags
BPF_CFLAGS := \
  -O2 \
  -g \
  -Wall \
  -target bpf \
  -D__TARGET_ARCH_$(ARCH) \
  -I./headers \
  -I./include \
  -I/usr/include

# All eBPF source files
BPF_SRCS := \
  maps/services.bpf.c \
  maps/policy.bpf.c \
  maps/flow.bpf.c \
  maps/routes.bpf.c \
  maps/peers.bpf.c \
  programs/xdp/sg_xdp.bpf.c \
  programs/tcx/sg_tcx.bpf.c \
  programs/socket/sg_sock_lb.bpf.c \
  programs/cgroup/sg_cgroup.bpf.c \
  programs/lsm/sg_lsm.bpf.c \
  programs/kprobe/sg_kprobe.bpf.c

BPF_OBJS := $(patsubst %.bpf.c,$(OUT_DIR)/%.bpf.o,$(BPF_SRCS))

.PHONY: all clean

all: $(OUT_DIR) $(BPF_OBJS)
	@echo "✓ eBPF compilation complete for $(ARCH)"

$(OUT_DIR):
	@mkdir -p \
	  $(OUT_DIR)/maps \
	  $(OUT_DIR)/programs/xdp \
	  $(OUT_DIR)/programs/tcx \
	  $(OUT_DIR)/programs/socket \
	  $(OUT_DIR)/programs/cgroup \
	  $(OUT_DIR)/programs/lsm \
	  $(OUT_DIR)/programs/kprobe

$(OUT_DIR)/%.bpf.o: %.bpf.c
	@echo "  CC  $<"
	$(CC) $(BPF_CFLAGS) -c $< -o $@

clean:
	rm -rf $(OUT_DIR)
