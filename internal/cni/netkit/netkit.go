// Package netkit provides Netkit-specific configuration options and definitions for CNI.
package netkit

// Mode defines the Netkit forwarding mode.
type Mode string

const (
	ModeL3 Mode = "l3"
	ModeL2 Mode = "l2"
)

// Policy defines the Netkit drop/forward policy.
type Policy string

const (
	PolicyForward   Policy = "forward"
	PolicyBlackhole Policy = "blackhole"
)

// Scrub defines packet header scrubbing policy.
type Scrub string

const (
	ScrubDefault Scrub = "default"
	ScrubNone    Scrub = "none"
)

// Config represents Netkit link creation parameters.
type Config struct {
	Mode       Mode
	Policy     Policy
	PeerPolicy Policy
	Scrub      Scrub
	MTU        int
}

// DefaultConfig returns optimal Netkit defaults for Kubernetes pod networking.
func DefaultConfig() Config {
	return Config{
		Mode:       ModeL3,
		Policy:     PolicyForward,
		PeerPolicy: PolicyForward,
		Scrub:      ScrubDefault,
		MTU:        1500,
	}
}
