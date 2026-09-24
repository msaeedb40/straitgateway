package ipam_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/straitgateway/straitgateway/internal/cni/ipam"
)

func TestStore(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "allocations.json")

	store := ipam.NewStore(storePath)

	state, err := store.Load()
	if err != nil {
		t.Fatalf("Load() failed on fresh store: %v", err)
	}

	state.Allocations["pod-1"] = ipam.PodAllocation{
		ContainerID:  "pod-1",
		PodName:      "nginx-test",
		PodNamespace: "default",
		IPv4:         "10.244.1.5/24",
		IPv6:         "fd00:10:244:1::5/64",
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	reloaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() failed after save: %v", err)
	}

	alloc, exists := reloaded.Allocations["pod-1"]
	if !exists {
		t.Fatalf("expected pod-1 in reloaded state")
	}
	if alloc.IPv4 != "10.244.1.5/24" || alloc.IPv6 != "fd00:10:244:1::5/64" {
		t.Errorf("got alloc %+v", alloc)
	}

	// Test Store CRUD helpers
	if err := store.Put(ipam.PodAllocation{
		ContainerID: "pod-2",
		IPv4:        "10.244.1.6/24",
	}); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	got2, exists2, err := store.Get("pod-2")
	if err != nil || !exists2 || got2.IPv4 != "10.244.1.6/24" {
		t.Errorf("Get pod-2 unexpected: exists=%v, val=%+v, err=%v", exists2, got2, err)
	}

	list, err := store.List()
	if err != nil || len(list) != 2 {
		t.Errorf("List unexpected: len=%d, err=%v", len(list), err)
	}

	if err := store.Delete("pod-1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	listAfter, err := store.List()
	if err != nil || len(listAfter) != 1 {
		t.Errorf("List after delete: len=%d, err=%v", len(listAfter), err)
	}
}

func TestStoreRestoreAllocator(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "allocations.json")

	store := ipam.NewStore(storePath)
	_ = store.Put(ipam.PodAllocation{
		ContainerID: "pod-persistent",
		IPv4:        "10.244.1.2/24", // base+1 is gw, base+2 would be next
	})

	alloc, err := ipam.NewAllocator([]string{"10.244.1.0/24"})
	if err != nil {
		t.Fatalf("NewAllocator: %v", err)
	}

	// Restore state into allocator
	if err := alloc.RestoreFromStore(store); err != nil {
		t.Fatalf("RestoreFromStore failed: %v", err)
	}

	// Because 10.244.1.2 was restored, next allocation must be 10.244.1.3
	nextIP, err := alloc.Allocate()
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}
	if nextIP != "10.244.1.3/24" {
		t.Errorf("expected 10.244.1.3/24, got %s", nextIP)
	}
}

func TestConcurrentStoreAccess(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "allocations.json")

	store := ipam.NewStore(storePath)
	const numGoroutines = 20

	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			cid := fmt.Sprintf("pod-%d", id)
			_ = store.Put(ipam.PodAllocation{
				ContainerID: cid,
				IPv4:        fmt.Sprintf("10.244.1.%d/24", id+10),
			})
			_, _, _ = store.Get(cid)
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(list) != numGoroutines {
		t.Errorf("expected %d allocations, got %d", numGoroutines, len(list))
	}
}

