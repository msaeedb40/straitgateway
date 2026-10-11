#!/usr/bin/env bash
# hack/kernel-matrix-fuzz.sh — Multi-kernel verifier fuzzing matrix runner
#
# Runs the Go native fuzzer against each supported kernel version.
# On the host kernel, tests are executed directly.  For other kernel versions,
# requires vmtest (github.com/cilium/ebpf/cmd/bpf2go/vmtest) or QEMU images.
#
# Usage:
#   ./hack/kernel-matrix-fuzz.sh [--fuzztime 30s] [--kernels 6.8,6.12,7.0]
#
# Environment:
#   FUZZTIME   - How long to fuzz each kernel (default: 30s)
#   VMTEST_BIN - Path to vmtest binary (optional; falls back to native only)

set -euo pipefail

FUZZTIME="${FUZZTIME:-30s}"
KERNELS="${KERNELS:-6.8,6.12,7.0}"

while [[ $# -gt 0 ]]; do
  case $1 in
    --fuzztime) FUZZTIME="$2"; shift 2 ;;
    --kernels)  KERNELS="$2";  shift 2 ;;
    *) echo "Unknown arg: $1"; exit 1 ;;
  esac
done

# Parse host kernel version
HOST_KERNEL=$(uname -r | grep -oP '^\d+\.\d+')
echo "==> StraitGateway Multi-Kernel Verifier Fuzzing Matrix"
echo "    Host kernel  : $(uname -r)"
echo "    Fuzz time    : $FUZZTIME"
echo "    Target matrix: $KERNELS"
echo ""

PASS=0
FAIL=0
SKIP=0

run_fuzz_native() {
  local kernel="$1"
  echo "  [fuzz] $kernel (native host) — go test -fuzz=FuzzVerifierMatrix -fuzztime=$FUZZTIME"
  if go test -v \
       -fuzz=FuzzVerifierMatrix \
       -fuzztime="$FUZZTIME" \
       ./test/fuzz/... 2>&1; then
    echo "  [fuzz] $kernel PASS"
    PASS=$((PASS+1))
  else
    echo "  [fuzz] $kernel FAIL"
    FAIL=$((FAIL+1))
  fi
}

run_unit_tests() {
  echo "  [unit] Running non-BPF unit tests (no root required)..."
  if go test -v -run "TestKernelVersionParsing|TestKernelMatrixCompleteness|TestBuiltinSeeds|TestMutateInstructions|TestCorpusAddAndRetrieve" \
       ./test/fuzz/... 2>&1; then
    echo "  [unit] PASS"
    PASS=$((PASS+1))
  else
    echo "  [unit] FAIL"
    FAIL=$((FAIL+1))
  fi
}

# Always run unit tests first
run_unit_tests

# Native kernel fuzzing
IFS=',' read -ra KERNEL_LIST <<< "$KERNELS"
for target in "${KERNEL_LIST[@]}"; do
  major_minor=$(echo "$target" | grep -oP '^\d+\.\d+')
  if [[ "$HOST_KERNEL" == "$major_minor" ]]; then
    run_fuzz_native "$target"
  else
    echo "  [fuzz] $target — SKIP (host kernel $HOST_KERNEL != $major_minor; use vmtest for cross-kernel)"
    SKIP=$((SKIP+1))
  fi
done

echo ""
echo "==> Fuzzing matrix results: PASS=$PASS FAIL=$FAIL SKIP=$SKIP"

if [[ $FAIL -gt 0 ]]; then
  exit 1
fi
exit 0
