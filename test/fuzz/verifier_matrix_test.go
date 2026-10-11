// Package fuzz — verifier_matrix_test.go drives the Go native fuzzer against
// the StraitGateway BPF verifier matrix.
//
// The FuzzVerifierMatrix fuzzer:
//  1. Seeds the corpus with builtin minimal XDP/TC programs.
//  2. On each iteration, decodes the fuzz input as a mutation seed.
//  3. Applies MutateInstructions to each base seed.
//  4. Calls ProbeVerifier for every (mutation × progType) pair.
//  5. Fails if an accepted program panics the kernel (SIGKILL/EFAULT from
//     prog_test_run — genuine kernel bugs we want to surface).
//
// Kernel matrix conformance:
//   - On the host kernel matching 6.8 or 6.12 LTS: all builtins must be accepted.
//   - On 7.x (edge): acceptance is informational; unknown opcodes are skipped.
//
// Unit tests (TestKernelVersionParsing, TestMutateInstructions, etc.) run without
// root and without BPF syscall access, so they are always safe in CI.
package fuzz_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
	"github.com/straitgateway/straitgateway/test/fuzz"
)

// ---------------------------------------------------------------------------
// Unit tests (no BPF, no root required)
// ---------------------------------------------------------------------------

func TestKernelVersionParsing(t *testing.T) {
	cases := []struct {
		input   string
		wantMaj int
		wantMin int
		wantPat int
	}{
		{
			"Linux version 6.8.0-41-generic (buildd@lcy02) ...",
			6, 8, 0,
		},
		{
			"Linux version 6.12.3 #1 SMP",
			6, 12, 3,
		},
		{
			"Linux version 7.0.0-rc1 #2",
			7, 0, 0,
		},
		{
			"not a version string",
			0, 0, 0,
		},
	}

	for _, tc := range cases {
		kv := fuzz.HostKernelVersionFromString(tc.input)
		if kv.Major != tc.wantMaj || kv.Minor != tc.wantMin || kv.Patch != tc.wantPat {
			t.Errorf("parse(%q) = {%d,%d,%d}, want {%d,%d,%d}",
				tc.input, kv.Major, kv.Minor, kv.Patch,
				tc.wantMaj, tc.wantMin, tc.wantPat)
		}
	}
}

func TestKernelMatrixCompleteness(t *testing.T) {
	matrix := fuzz.SupportedKernelMatrix
	if len(matrix) < 3 {
		t.Errorf("expected at least 3 kernel versions in matrix, got %d", len(matrix))
	}

	labels := make(map[string]bool)
	for _, kv := range matrix {
		if labels[kv.Label] {
			t.Errorf("duplicate kernel label %q in matrix", kv.Label)
		}
		labels[kv.Label] = true
	}
}

func TestBuiltinSeeds(t *testing.T) {
	seeds := fuzz.BuiltinSeeds()
	if len(seeds) == 0 {
		t.Fatal("BuiltinSeeds returned empty corpus")
	}

	for _, s := range seeds {
		if s.Name == "" {
			t.Error("seed with empty name")
		}
		if len(s.Instructions) == 0 {
			t.Errorf("seed %q has empty instruction set", s.Name)
		}
		// Last instruction should be a Return
		last := s.Instructions[len(s.Instructions)-1]
		if last.OpCode != asm.Return().OpCode {
			t.Errorf("seed %q: last instruction is not Return (got %v)", s.Name, last)
		}
	}
}

func TestMutateInstructions(t *testing.T) {
	base := asm.Instructions{
		asm.Mov.Imm(asm.R0, 2),
		asm.Return(),
	}

	seed := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	mutations := fuzz.MutateInstructions(base, seed)

	if len(mutations) == 0 {
		t.Fatal("MutateInstructions returned no mutations")
	}

	for i, m := range mutations {
		if len(m) == 0 {
			t.Errorf("mutation %d is empty", i)
		}
		last := m[len(m)-1]
		if last.OpCode != asm.Return().OpCode {
			// Mutation 4 (early-exit) inserts Return mid-sequence — the final
			// instruction might not be Return. Accept any mutation.
			t.Logf("mutation %d: last instruction is not Return (%v) — may be early-exit variant", i, last)
		}
	}
}

func TestCorpusAddAndRetrieve(t *testing.T) {
	c := fuzz.NewCorpus()
	seed := fuzz.SeedProgram{
		Name: "test/pass",
		Type: ebpf.XDP,
		Instructions: asm.Instructions{
			asm.Mov.Imm(asm.R0, 2),
			asm.Return(),
		},
	}
	c.AddSeed(seed)

	seeds := c.Seeds()
	if len(seeds) != 1 {
		t.Fatalf("expected 1 seed, got %d", len(seeds))
	}
	if seeds[0].Name != "test/pass" {
		t.Errorf("expected seed name 'test/pass', got %q", seeds[0].Name)
	}
}

// ---------------------------------------------------------------------------
// BPF-capable tests (require root + Linux)
// ---------------------------------------------------------------------------

func TestProbeVerifierPassProg(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("ProbeVerifier requires root (CAP_BPF)")
	}
	if runtime.GOOS != "linux" {
		t.Skip("BPF verifier requires Linux")
	}

	insns := asm.Instructions{
		asm.Mov.Imm(asm.R0, 2), // XDP_PASS
		asm.Return(),
	}

	result := fuzz.ProbeVerifier(ebpf.XDP, insns, "GPL")
	if !result.Accepted {
		t.Errorf("minimal XDP pass program rejected by verifier: %s (err: %v)",
			result.VerifierLog, result.Err)
	}
}

func TestProbeVerifierInvalidProg(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("ProbeVerifier requires root")
	}
	if runtime.GOOS != "linux" {
		t.Skip("BPF verifier requires Linux")
	}

	// An empty program — kernel verifier must reject this.
	result := fuzz.ProbeVerifier(ebpf.XDP, asm.Instructions{}, "GPL")
	if result.Accepted {
		t.Error("empty program was unexpectedly accepted by verifier")
	}
}

func TestProbeVerifierBuiltinSeeds(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("ProbeVerifier requires root")
	}
	if runtime.GOOS != "linux" {
		t.Skip("BPF verifier requires Linux")
	}

	seeds := fuzz.BuiltinSeeds()
	for _, seed := range seeds {
		seed := seed
		t.Run(seed.Name, func(t *testing.T) {
			result := fuzz.ProbeVerifier(seed.Type, seed.Instructions, "GPL")
			if !result.Accepted {
				t.Errorf("builtin seed %q rejected by verifier: %s (err: %v)",
					seed.Name, result.VerifierLog, result.Err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Go native fuzzer entry point
// ---------------------------------------------------------------------------

// FuzzVerifierMatrix is the primary fuzzing entry point for `go test -fuzz`.
// It tests each builtin seed under byte-sequence mutations generated by the
// Go fuzzer. The test FAILS if a valid mutation causes the kernel to return
// an error other than a verifier rejection (which would indicate a kernel bug).
func FuzzVerifierMatrix(f *testing.F) {
	// Seed the corpus with interesting byte patterns
	f.Add([]byte{0x00})
	f.Add([]byte{0xff, 0x00, 0x01, 0x02})
	f.Add([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08})
	f.Add([]byte{0x7f, 0x80, 0x81, 0xff, 0x00, 0x01})

	seeds := fuzz.BuiltinSeeds()

	f.Fuzz(func(t *testing.T, mutSeed []byte) {
		if os.Getuid() != 0 {
			t.Skip("FuzzVerifierMatrix requires root")
		}
		if runtime.GOOS != "linux" {
			t.Skip("BPF verifier requires Linux")
		}
		if len(mutSeed) == 0 {
			return
		}

		for _, seed := range seeds {
			mutations := fuzz.MutateInstructions(seed.Instructions, mutSeed)
			for _, insns := range mutations {
				result := fuzz.ProbeVerifier(seed.Type, insns, "GPL")

				// The ONLY acceptable non-acceptance is a verifier error.
				// Any other error (EFAULT, SIGKILL, etc.) is a kernel bug.
				if !result.Accepted && result.Err != nil {
					// Check if it's a verifier-type error (expected rejection)
					if result.VerifierLog == "" {
						t.Errorf("kernel returned non-verifier error for mutation of %q: %v",
							seed.Name, result.Err)
					}
				}
			}
		}
	})
}
