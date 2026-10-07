package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/straitgateway/straitgateway/internal/agent"
	"github.com/straitgateway/straitgateway/pkg/version"
)

func main() {
	var nodeName string
	var nodeIP string
	var localASN uint

	flag.StringVar(&nodeName, "node-name", "", "Name of the Kubernetes node")
	flag.StringVar(&nodeIP, "node-ip", "127.0.0.1", "Primary IP address of the node")
	flag.UintVar(&localASN, "asn", 65001, "Local BGP Autonomous System Number")
	flag.Parse()

	if nodeName == "" {
		host, _ := os.Hostname()
		nodeName = host
	}

	fmt.Printf("Starting straitd (StraitGateway Node Daemon) - %s\n", version.Info())

	cfg := agent.Config{
		NodeName: nodeName,
		NodeIP:   net.ParseIP(nodeIP),
		LocalASN: uint32(localASN),
	}

	mgr, err := agent.NewManager(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize straitd: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := mgr.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "straitd exited with error: %v\n", err)
		os.Exit(1)
	}
}
