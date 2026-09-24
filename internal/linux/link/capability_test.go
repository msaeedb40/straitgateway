package link_test

import (
	"testing"

	"github.com/straitgateway/straitgateway/internal/linux/link"
)

func TestDetectCapabilities(t *testing.T) {
	caps := link.DetectCapabilities()
	if caps == nil {
		t.Fatal("DetectCapabilities() returned nil")
	}

	if caps.Major == 0 {
		t.Errorf("expected major kernel version > 0, got %d", caps.Major)
	}

	if caps.DatapathMode != link.DatapathModeNetkit && caps.DatapathMode != link.DatapathModeVeth {
		t.Errorf("unexpected datapath mode %v", caps.DatapathMode)
	}
}
