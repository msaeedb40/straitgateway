// Package types defines shared types used across StraitGateway packages.
package types

// ServiceType represents the Kubernetes Service exposure type.
type ServiceType string

const (
	ServiceTypeClusterIP    ServiceType = "ClusterIP"
	ServiceTypeNodePort     ServiceType = "NodePort"
	ServiceTypeLoadBalancer ServiceType = "LoadBalancer"
)

// Protocol represents a network protocol.
type Protocol string

const (
	ProtocolTCP  Protocol = "TCP"
	ProtocolUDP  Protocol = "UDP"
	ProtocolSCTP Protocol = "SCTP"
)

// Phase represents a component lifecycle phase.
type Phase string

const (
	PhaseInitializing Phase = "Initializing"
	PhaseRunning      Phase = "Running"
	PhaseDegraded     Phase = "Degraded"
	PhaseError        Phase = "Error"
	PhaseStopped      Phase = "Stopped"
)

// TopologyMode is the transit gateway topology.
type TopologyMode string

const (
	TopologyHubSpoke   TopologyMode = "hub-spoke"
	TopologyMesh       TopologyMode = "mesh"
	TopologyPeerToPeer TopologyMode = "peer-to-peer"
	TopologyHubToHub   TopologyMode = "hub-to-hub"
	TopologyHybrid     TopologyMode = "hybrid"
)
