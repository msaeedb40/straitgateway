// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package main implements the straitgatewayd node agent.
//
// straitgatewayd runs as a privileged DaemonSet on every node.
// Responsibilities:
//   - Serve gRPC IPAM and endpoint registration API (Unix socket)
//   - Load and attach eBPF programs (NetKit TCX, XDP, sockops, LSM)
//   - Pin BPF maps to /sys/fs/bpf/straitgateway/
//   - Reconcile dataplane state from the controller via NodeNetworkConfig CRD
//   - Manage WireGuard tunnel interfaces for transit encryption
//   - Handle CNI plugin calls via Unix socket
package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/msaeedb40/straitgateway/internal/version"
	"github.com/msaeedb40/straitgateway/platform/process"
)

const (
	defaultSocketPath = "/run/straitgateway/daemon.sock"
	defaultBPFFS      = "/sys/fs/bpf/straitgateway"
)

func main() {
	var (
		socketPath string
		bpfFSPath  string
		logLevel   string
		nodeName   string
	)

	flag.StringVar(&socketPath, "socket", defaultSocketPath,
		"Unix socket path for CNI and controller gRPC connections.")
	flag.StringVar(&bpfFSPath, "bpffs", defaultBPFFS,
		"BPF filesystem mount path for pinned maps and programs.")
	flag.StringVar(&logLevel, "log-level", "info",
		"Log level: debug, info, warn, error.")
	flag.StringVar(&nodeName, "node-name", os.Getenv("NODE_NAME"),
		"Kubernetes node name (injected via downward API).")
	flag.Parse()

	// Configure structured logger.
	var zapCfg zap.Config
	if logLevel == "debug" {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}
	log, err := zapCfg.Build()
	if err != nil {
		panic(err)
	}
	defer log.Sync() //nolint:errcheck

	log.Info("starting straitgatewayd",
		zap.String("version", version.Version),
		zap.String("commit", version.Commit),
		zap.String("node", nodeName),
	)

	// Verify platform requirements.
	if err := process.CheckKernelVersion(log); err != nil {
		log.Fatal("kernel version check failed", zap.Error(err))
	}
	if err := process.EnsureBPFFS(bpfFSPath); err != nil {
		log.Fatal("BPF filesystem setup failed", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Create Unix socket directory.
	if err := os.MkdirAll("/run/straitgateway", 0750); err != nil {
		log.Fatal("creating socket directory", zap.Error(err))
	}
	_ = os.Remove(socketPath)

	var lc net.ListenConfig
	lis, err := lc.Listen(ctx, "unix", socketPath)
	if err != nil {
		log.Fatal("listening on socket", zap.String("path", socketPath), zap.Error(err))
	}

	// Create gRPC server for CNI and controller communication.
	grpcSrv := grpc.NewServer(
		grpc.MaxRecvMsgSize(4<<20),
		grpc.MaxSendMsgSize(4<<20),
	)

	// TODO: Register agent service implementation:
	// agentpb.RegisterAgentServer(grpcSrv, newAgentServer(log, bpfFSPath, nodeName))

	// Start gRPC server.
	go func() {
		log.Info("gRPC server listening", zap.String("socket", socketPath))
		if err := grpcSrv.Serve(lis); err != nil {
			log.Error("gRPC server error", zap.Error(err))
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()
	log.Info("received shutdown signal, stopping straitgatewayd")
	grpcSrv.GracefulStop()
	log.Info("straitgatewayd stopped")
}
