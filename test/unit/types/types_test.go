package types_test

import (
	"testing"

	"github.com/straitgateway/straitgateway/pkg/types"
)

func TestPhaseConstants(t *testing.T) {
	phases := []types.Phase{
		types.PhaseInitializing,
		types.PhaseRunning,
		types.PhaseDegraded,
		types.PhaseError,
		types.PhaseStopped,
	}
	for _, p := range phases {
		if string(p) == "" {
			t.Errorf("phase constant must not be empty")
		}
	}
}

func TestTopologyModeConstants(t *testing.T) {
	modes := []types.TopologyMode{
		types.TopologyHubSpoke,
		types.TopologyMesh,
		types.TopologyPeerToPeer,
		types.TopologyHubToHub,
		types.TopologyHybrid,
	}
	for _, m := range modes {
		if string(m) == "" {
			t.Errorf("topology mode constant must not be empty")
		}
	}
}
