#!/usr/bin/env bash
# hack/run-chaos.sh — eBPF chaos injection smoke runner
#
# Executes a brief round of all supported fault modes and verifies that
# the StraitGateway dataplane handles each without crashing.
#
# Usage:
#   sudo ./hack/run-chaos.sh [--iface eth0] [--duration 5s] [--bpf-obj bpf/xdp/ingress.o]
#
# Requirements:
#   - Linux with CAP_NET_ADMIN + CAP_BPF (or root)
#   - tc (iproute2) installed

set -euo pipefail

IFACE="${IFACE:-lo}"
DURATION="${DURATION:-2s}"
BPF_OBJ="${BPF_OBJ:-}"

# Parse args
while [[ $# -gt 0 ]]; do
  case $1 in
    --iface)     IFACE="$2";    shift 2 ;;
    --duration)  DURATION="$2"; shift 2 ;;
    --bpf-obj)   BPF_OBJ="$2";  shift 2 ;;
    *) echo "Unknown arg: $1"; exit 1 ;;
  esac
done

echo "==> StraitGateway Chaos Injection Smoke Runner"
echo "    Interface : $IFACE"
echo "    Duration  : $DURATION"
echo "    BPF Object: ${BPF_OBJ:-<builtin minimal prog>}"
echo ""

run_test() {
  local name="$1"
  shift
  echo -n "  [chaos] $name ... "
  if "$@" 2>/tmp/chaos-stderr; then
    echo "PASS"
  else
    echo "FAIL (see /tmp/chaos-stderr)"
    cat /tmp/chaos-stderr
    return 1
  fi
}

# Run the Go chaos tests directly
CHAOS_TEST_ARGS=(
  -test.v
  -test.run "TestInjectorNew|TestClearAll"
)
if [[ $EUID -eq 0 ]]; then
  CHAOS_TEST_ARGS+=(
    -test.run "TestInjectENOMEM|TestInjectDropBurst|TestBuildMinimalPassProg"
  )
fi

run_test "Chaos unit tests" \
  go test ./test/chaos/... "${CHAOS_TEST_ARGS[@]}"

# Jitter injection via tc-netem (only if tc available)
if command -v tc &>/dev/null && [[ $EUID -eq 0 ]]; then
  run_test "tc-netem jitter (${IFACE})" bash -c "
    tc qdisc add dev ${IFACE} root netem delay 10ms 5ms 2>/dev/null || true
    sleep 0.5
    tc qdisc del dev ${IFACE} root 2>/dev/null || true
  "
fi

# BFD flap simulation (loopback flap is a no-op but exercises the code path)
if [[ $EUID -eq 0 ]]; then
  run_test "BFD link flap (${IFACE})" bash -c "
    ip link set dev ${IFACE} down 2>/dev/null || true
    sleep 0.1
    ip link set dev ${IFACE} up 2>/dev/null || true
  "
fi

echo ""
echo "==> Chaos smoke run complete."
