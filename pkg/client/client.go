// Package client provides a Go client for the StraitGateway internal API.
// Used by sgctl, tgwd, and test tooling to communicate with straitd.
package client

import "fmt"

// Client is a StraitGateway internal API client.
type Client struct {
	addr string
}

// New creates a new Client connected to the given address.
// addr examples: "unix:///run/straitd/api.sock" or "tcp://localhost:9099"
func New(addr string) (*Client, error) {
	if addr == "" {
		return nil, fmt.Errorf("client.New: address must not be empty")
	}
	return &Client{addr: addr}, nil
}

// Close closes the client connection.
func (c *Client) Close() error { return nil }

// Ping checks connectivity to straitd.
func (c *Client) Ping() error {
	// TODO: implement gRPC Health Check
	return fmt.Errorf("Ping: not yet implemented")
}
