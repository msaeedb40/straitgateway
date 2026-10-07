package loadbalancer

import (
	"net"
)

// EndpointState indicates backend health
type EndpointState uint8

const (
	EndpointStateActive   EndpointState = 1
	EndpointStateDraining EndpointState = 2
	EndpointStateOffline  EndpointState = 3
)

// Backend represents an individual service pod backend
type Backend struct {
	ID       uint32
	IP       net.IP
	Port     uint16
	Weight   uint8
	State    EndpointState
	NodeName string
}
