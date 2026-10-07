package loader

import (
	"path/filepath"
	"testing"

	"github.com/straitgateway/straitgateway/internal/datapath/maps"
)

func TestLoaderWithCompiledObjects(t *testing.T) {
	tcIngressObj := filepath.Join("..", "..", "..", "bpf", "tc", "ingress.o")

	prog, err := LoadFromFile(tcIngressObj, "tc_ingress_func")
	if err != nil {
		t.Logf("Skipping live BPF program load if kernel unprivileged or without CAP_BPF: %v", err)
		return
	}
	defer prog.Close()

	if prog.Type() == 0 {
		t.Error("expected non-zero program type")
	}

	cfg := maps.DefaultConfig()
	mm, err := maps.NewMapManager(cfg)
	if err != nil {
		t.Fatalf("failed to init MapManager: %v", err)
	}
	defer mm.Close()

	l := NewLoader(mm)
	if l == nil {
		t.Fatal("expected non-nil loader")
	}

	dc := &DataplaneCollection{
		mapManager: mm,
		tcIngress:  prog,
	}
	if err := dc.Close(); err != nil {
		t.Fatalf("failed to close dataplane collection: %v", err)
	}
}

func TestInterfaceIndexNotFound(t *testing.T) {
	_, err := GetInterfaceIndexByName("non-existent-sg-if0")
	if err == nil {
		t.Error("expected error for non-existent interface, got nil")
	}
}
