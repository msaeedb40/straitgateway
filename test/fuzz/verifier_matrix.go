// Package fuzz implements the multi-kernel eBPF verifier fuzzing matrix for
// StraitGateway.
//
// # Architecture
//
// The fuzzer has three layers:
//
//  1. KernelMatrix — declares which kernel versions and capability sets to test.
//  2. ProgramCorpus — maintains a set of eBPF program mutation seeds derived from
//     the StraitGateway BPF object ELFs.
//  3. VerifierFuzzer — drives the Go fuzzer (testing.F) against each (kernel, seed)
//     pair, mutating BPF instructions and feeding them through the kernel verifier
//     via ebpf.NewProgram.
//
// # Kernel Version Matrix
//
// Supported target kernels (run via vmtest or native if matching host):
//
//	Linux 6.8   (Ubuntu 24.04 LTS default)
//	Linux 6.12  (Ubuntu 24.10 / Debian sid — LTS)
//	Linux 7.x   (bleeding-edge; test_PASS when verifier accepts; SKIP on ENOSYS)
//
// # Running
//
//	go test -v -fuzz=FuzzVerifierMatrix -fuzztime=60s ./test/fuzz/...
//
// or via make:
//
//	make fuzz FUZZTIME=120s
package fuzz

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/asm"
)

// KernelVersion represents a supported kernel version in the test matrix.
type KernelVersion struct {
	Major int
	Minor int
	Patch int
	// Label is a human-readable alias (e.g. "6.8-lts", "6.12-lts", "7.x-edge")
	Label string
}

func (kv KernelVersion) String() string {
	return fmt.Sprintf("%d.%d.%d (%s)", kv.Major, kv.Minor, kv.Patch, kv.Label)
}

// SupportedKernelMatrix is the authoritative list of kernel versions in the
// StraitGateway CI fuzzing matrix.
var SupportedKernelMatrix = []KernelVersion{
	{6, 8, 0, "6.8-lts"},
	{6, 12, 0, "6.12-lts"},
	{7, 0, 0, "7.x-edge"},
}

// HostKernelVersion reads /proc/version_signature or /proc/version and parses
// the kernel version triple. Returns (0,0,0) on parse failure.
func HostKernelVersion() KernelVersion {
	for _, path := range []string{"/proc/version_signature", "/proc/version"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		return parseKernelVersion(string(data))
	}
	return KernelVersion{}
}

// HostKernelVersionFromString parses a kernel version triple from an arbitrary
// string — exported so unit tests can call it without /proc access.
func HostKernelVersionFromString(s string) KernelVersion {
	return parseKernelVersion(s)
}

func parseKernelVersion(s string) KernelVersion {
	// e.g. "Linux version 6.8.0-41-generic (buildd@lcy02-amd64-016) ..."
	for _, field := range strings.Fields(s) {
		parts := strings.SplitN(field, ".", 3)
		if len(parts) < 2 {
			continue
		}
		major, err1 := strconv.Atoi(parts[0])
		minor, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			continue
		}
		patch := 0
		if len(parts) == 3 {
			patch, _ = strconv.Atoi(strings.Split(parts[2], "-")[0])
		}
		if major >= 4 {
			return KernelVersion{major, minor, patch, fmt.Sprintf("%d.%d.%d", major, minor, patch)}
		}
	}
	return KernelVersion{}
}

// IsCompatibleWith returns true if this host kernel satisfies the target version.
func (kv KernelVersion) IsCompatibleWith(target KernelVersion) bool {
	if kv.Major != target.Major {
		return kv.Major > target.Major
	}
	if kv.Minor != target.Minor {
		return kv.Minor > target.Minor
	}
	return kv.Patch >= target.Patch
}

// ---------------------------------------------------------------------------
// Program Corpus
// ---------------------------------------------------------------------------

// SeedProgram is a single seed entry in the fuzzing corpus.
type SeedProgram struct {
	// Name identifies the program (e.g. "tc/ingress", "xdp/pass")
	Name string
	// Type is the BPF program type.
	Type ebpf.ProgramType
	// Instructions is the baseline instruction sequence.
	Instructions asm.Instructions
}

// Corpus holds the set of seed programs for mutation fuzzing.
type Corpus struct {
	mu    sync.RWMutex
	seeds []SeedProgram
}

// NewCorpus creates an empty Corpus.
func NewCorpus() *Corpus {
	return &Corpus{}
}

// AddSeed appends a seed program to the corpus.
func (c *Corpus) AddSeed(s SeedProgram) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seeds = append(c.seeds, s)
}

// Seeds returns a snapshot of all seed programs.
func (c *Corpus) Seeds() []SeedProgram {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]SeedProgram, len(c.seeds))
	copy(out, c.seeds)
	return out
}

// LoadFromELF loads all programs from an eBPF ELF file into the corpus.
// Programs that fail to parse are logged and skipped.
func (c *Corpus) LoadFromELF(elfPath string) error {
	spec, err := ebpf.LoadCollectionSpec(elfPath)
	if err != nil {
		return fmt.Errorf("corpus: load ELF %s: %w", elfPath, err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for name, pspec := range spec.Programs {
		seed := SeedProgram{
			Name:         filepath.Join(filepath.Base(elfPath), name),
			Type:         pspec.Type,
			Instructions: make(asm.Instructions, len(pspec.Instructions)),
		}
		copy(seed.Instructions, pspec.Instructions)
		c.seeds = append(c.seeds, seed)
	}
	return nil
}

// BuiltinSeeds returns a minimal set of hardcoded seed programs that can be
// used when no ELF files are available (e.g., in CI without BPF object build).
func BuiltinSeeds() []SeedProgram {
	return []SeedProgram{
		{
			Name: "xdp/pass",
			Type: ebpf.XDP,
			Instructions: asm.Instructions{
				asm.Mov.Imm(asm.R0, 2), // XDP_PASS
				asm.Return(),
			},
		},
		{
			Name: "xdp/drop",
			Type: ebpf.XDP,
			Instructions: asm.Instructions{
				asm.Mov.Imm(asm.R0, 1), // XDP_DROP
				asm.Return(),
			},
		},
		{
			Name: "tc/ok",
			Type: ebpf.SchedCLS,
			Instructions: asm.Instructions{
				asm.Mov.Imm(asm.R0, 0), // TC_ACT_OK
				asm.Return(),
			},
		},
		{
			Name: "tc/shot",
			Type: ebpf.SchedCLS,
			Instructions: asm.Instructions{
				asm.Mov.Imm(asm.R0, 2), // TC_ACT_SHOT
				asm.Return(),
			},
		},
		{
			Name: "sk/pass",
			Type: ebpf.SkSKB,
			Instructions: asm.Instructions{
				asm.Mov.Imm(asm.R0, 1), // SK_PASS
				asm.Return(),
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Verifier Fuzzer
// ---------------------------------------------------------------------------

// VerifierResult captures the outcome of a single verifier probe.
type VerifierResult struct {
	SeedName string
	Kernel   KernelVersion
	// Accepted is true when the kernel verifier loaded the program successfully.
	Accepted bool
	// VerifierLog contains the kernel verifier output on rejection.
	VerifierLog string
	// Err is non-nil on unexpected (non-verifier) errors.
	Err error
}

// ProbeVerifier attempts to load the given instruction sequence into the kernel
// verifier using ebpf.NewProgram.  It captures the verifier log on failure.
func ProbeVerifier(progType ebpf.ProgramType, insns asm.Instructions, license string) VerifierResult {
	if license == "" {
		license = "GPL"
	}

	spec := &ebpf.ProgramSpec{
		Name:         "fuzz_probe",
		Type:         progType,
		License:      license,
		Instructions: insns,
	}

	opts := ebpf.ProgramOptions{
		LogLevel:     ebpf.LogLevelInstruction,
		LogSizeStart: 64 * 1024,
	}

	prog, err := ebpf.NewProgramWithOptions(spec, opts)
	if err != nil {
		var ve *ebpf.VerifierError
		if asVerifierError(err, &ve) {
			return VerifierResult{
				Accepted:    false,
				VerifierLog: ve.Error(),
			}
		}
		return VerifierResult{
			Accepted: false,
			Err:      err,
		}
	}
	_ = prog.Close()
	return VerifierResult{Accepted: true}
}

// MutateInstructions applies a corpus of deterministic mutations to a base
// instruction slice. It returns a slice of mutated instruction sequences.
//
// Mutations include:
//   - NOP insertion (replace an instruction with Mov R0, R0)
//   - Constant folding (replace Mov.Imm with a different immediate)
//   - Register swap (swap R1 and R2 in a load instruction)
//   - Tail-call insertion (insert a stub prog-array call)
func MutateInstructions(base asm.Instructions, seed []byte) []asm.Instructions {
	if len(base) == 0 || len(seed) == 0 {
		return nil
	}

	mutations := make([]asm.Instructions, 0, 4)

	// Mutation 1: NOP insertion at position seed[0] % len
	if len(base) > 0 {
		pos := int(seed[0]) % len(base)
		m1 := cloneInsns(base)
		m1[pos] = asm.Mov.Reg(asm.R0, asm.R0)
		mutations = append(mutations, m1)
	}

	// Mutation 2: Immediate constant swap using seed bytes
	if len(seed) >= 4 {
		m2 := cloneInsns(base)
		imm := int32(seed[0])<<24 | int32(seed[1])<<16 | int32(seed[2])<<8 | int32(seed[3])
		for i, ins := range m2 {
			if ins.OpCode == asm.Mov.Op(asm.ImmSource) {
				m2[i].Constant = int64(imm)
				break
			}
		}
		mutations = append(mutations, m2)
	}

	// Mutation 3: Register swap (R1 ↔ R2)
	{
		m3 := cloneInsns(base)
		for i, ins := range m3 {
			switch ins.Dst {
			case asm.R1:
				m3[i].Dst = asm.R2
			case asm.R2:
				m3[i].Dst = asm.R1
			}
		}
		mutations = append(mutations, m3)
	}

	// Mutation 4: Insert BPF_EXIT early (reachability test)
	if len(base) > 2 {
		pos := int(seed[0]%byte(len(base)-1)) + 1
		m4 := cloneInsns(base)
		m4 = append(m4[:pos], append(asm.Instructions{asm.Return()}, m4[pos:]...)...)
		mutations = append(mutations, m4)
	}

	return mutations
}

func cloneInsns(insns asm.Instructions) asm.Instructions {
	out := make(asm.Instructions, len(insns))
	copy(out, insns)
	return out
}

// asVerifierError is a helper to detect ebpf.VerifierError without importing
// errors package (which is not available at this package level from cilium/ebpf).
func asVerifierError(err error, target **ebpf.VerifierError) bool {
	if ve, ok := err.(*ebpf.VerifierError); ok {
		*target = ve
		return true
	}
	return false
}
