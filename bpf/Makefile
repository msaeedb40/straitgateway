# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 straitgateway Authors

CLANG ?= $(shell which clang-22 2>/dev/null || which clang 2>/dev/null || echo clang)
LLVM_STRIP ?= $(shell which llvm-strip-22 2>/dev/null || which llvm-strip 2>/dev/null || echo llvm-strip)
BPFTOOL ?= bpftool

ARCH ?= $(shell uname -m | sed 's/x86_64/x86/' | sed 's/aarch64/arm64/')
MULTIARCH ?= $(shell gcc -print-multiarch 2>/dev/null || echo x86_64-linux-gnu)

CFLAGS := -g -O2 -target bpf \
	-D__TARGET_ARCH_$(ARCH) \
	-Wall \
	-I. -I./maps -I/usr/include -I/usr/include/$(MULTIARCH) -I/usr/include/bpf

BPF_SRCS := $(wildcard cni/*.c) \
	$(wildcard dataplane/*.c) \
	$(wildcard service/*.c) \
	$(wildcard nat/*.c) \
	$(wildcard policy/*.c) \
	$(wildcard observability/*.c)

BPF_OBJS := $(BPF_SRCS:.c=.o)

.PHONY: all clean

all: $(BPF_OBJS)

%.o: %.c
	@if which $(CLANG) >/dev/null 2>&1; then \
		echo "Compiling BPF object $@ with $(CLANG)..."; \
		$(CLANG) $(CFLAGS) -c $< -o $@ || exit 1; \
		if which $(LLVM_STRIP) >/dev/null 2>&1; then \
			$(LLVM_STRIP) -g $@ || true; \
		fi; \
	else \
		echo "Warning: no clang compiler found, skipping $@"; \
	fi

clean:
	rm -f $(BPF_OBJS)
