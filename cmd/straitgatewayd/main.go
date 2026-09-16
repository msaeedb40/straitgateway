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
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/msaeedb40/straitgateway/ebpf/loader"
	"github.com/msaeedb40/straitgateway/identity"
	"github.com/msaeedb40/straitgateway/internal/version"
	"github.com/msaeedb40/straitgateway/ipam"
	"github.com/msaeedb40/straitgateway/pkg/api"
	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
	"github.com/msaeedb40/straitgateway/platform/process"
	agentv1 "github.com/msaeedb40/straitgateway/proto/agent/v1"
	commonv1 "github.com/msaeedb40/straitgateway/proto/common/v1"
)

const (
	defaultSocketPath = "/run/straitgateway/daemon.sock"
	defaultBPFFS      = "/sys/fs/bpf/straitgateway"
	defaultHealthAddr = ":8084"
)

type agentServer struct {
	agentv1.UnimplementedAgentServiceServer
	log         *zap.Logger
	nodeName    string
	bpfFSPath   string
	ipamAlloc   *ipam.Allocator
	idAlloc     *identity.Allocator
	endpointsMu sync.Mutex
	endpoints   map[string]*endpointInfo
}

type endpointInfo struct {
	containerID string
	ip          netip.Addr
	ifIndex     int32
	identity    sgtypes.Identity
}

func newAgentServer(log *zap.Logger, nodeName, bpfFSPath string, podCIDR netip.Prefix) (*agentServer, error) {
	alloc, err := ipam.New(podCIDR)
	if err != nil {
		return nil, fmt.Errorf("initializing IPAM for CIDR %s: %w", podCIDR, err)
	}

	return &agentServer{
		log:       log,
		nodeName:  nodeName,
		bpfFSPath: bpfFSPath,
		ipamAlloc: alloc,
		idAlloc:   identity.NewAllocator(),
		endpoints: make(map[string]*endpointInfo),
	}, nil
}

func (s *agentServer) AllocateIP(ctx context.Context, req *agentv1.AllocateIPRequest) (*agentv1.AllocateIPResponse, error) {
	s.log.Info("allocating IP for container",
		zap.String("containerID", req.ContainerId),
		zap.String("namespace", req.Namespace),
		zap.String("pod", req.PodName),
	)

	ip, err := s.ipamAlloc.Allocate()
	if err != nil {
		return nil, fmt.Errorf("ipam allocation failed: %w", err)
	}

	gw := s.ipamAlloc.Gateway()
	cidr := s.ipamAlloc.CIDR()

	gwIP := gw.As4()
	allocIP := ip.As4()

	gwUint := binary.BigEndian.Uint32(gwIP[:])
	allocUint := binary.BigEndian.Uint32(allocIP[:])

	resp := &agentv1.AllocateIPResponse{
		Ipv4Address: &commonv1.IPAddress{
			Address: &commonv1.IPAddress_Ipv4{Ipv4: allocUint},
		},
		Gateway: &commonv1.IPAddress{
			Address: &commonv1.IPAddress_Ipv4{Ipv4: gwUint},
		},
		Ipv4Cidr: &commonv1.IPPrefix{
			Address: &commonv1.IPAddress{
				Address: &commonv1.IPAddress_Ipv4{Ipv4: allocUint},
			},
			PrefixLength: uint32(cidr.Bits()),
		},
		Mtu: 1500,
	}

	s.endpointsMu.Lock()
	s.endpoints[req.ContainerId] = &endpointInfo{
		containerID: req.ContainerId,
		ip:          ip,
	}
	s.endpointsMu.Unlock()

	return resp, nil
}

func (s *agentServer) ReleaseIP(ctx context.Context, req *agentv1.ReleaseIPRequest) (*agentv1.ReleaseIPResponse, error) {
	s.endpointsMu.Lock()
	defer s.endpointsMu.Unlock()

	if ep, ok := s.endpoints[req.ContainerId]; ok {
		s.ipamAlloc.Release(ep.ip)
		delete(s.endpoints, req.ContainerId)
		s.log.Info("released IP for container", zap.String("containerID", req.ContainerId), zap.String("ip", ep.ip.String()))
	}

	return &agentv1.ReleaseIPResponse{}, nil
}

func (s *agentServer) ConfigureEndpoint(ctx context.Context, req *agentv1.ConfigureEndpointRequest) (*agentv1.ConfigureEndpointResponse, error) {
	s.endpointsMu.Lock()
	defer s.endpointsMu.Unlock()

	labels := make(map[string]string)
	for k, v := range req.Labels {
		labels[k] = v
	}
	labels["namespace"] = req.Namespace

	entry := s.idAlloc.Allocate(labels, 0) // default backbone segment 0
	secID := entry.Identity

	if ep, ok := s.endpoints[req.ContainerId]; ok {
		ep.ifIndex = req.HostIfindex
		ep.identity = secID
	}

	s.log.Info("configured endpoint",
		zap.String("containerID", req.ContainerId),
		zap.Int32("hostIfindex", req.HostIfindex),
		zap.Uint32("identity", uint32(secID)),
	)

	return &agentv1.ConfigureEndpointResponse{
		Identity: &commonv1.Identity{
			Id:         uint32(secID),
			LabelsHash: entry.LabelsHash,
		},
	}, nil
}

func (s *agentServer) DeleteEndpoint(ctx context.Context, req *agentv1.DeleteEndpointRequest) (*agentv1.DeleteEndpointResponse, error) {
	s.endpointsMu.Lock()
	defer s.endpointsMu.Unlock()

	if ep, ok := s.endpoints[req.ContainerId]; ok {
		s.ipamAlloc.Release(ep.ip)
		delete(s.endpoints, req.ContainerId)
		s.log.Info("deleted endpoint", zap.String("containerID", req.ContainerId))
	}

	return &agentv1.DeleteEndpointResponse{}, nil
}

func (s *agentServer) GetStatus(ctx context.Context, req *agentv1.GetStatusRequest) (*agentv1.GetStatusResponse, error) {
	kv, _ := process.GetRunningKernelVersion()
	return &agentv1.GetStatusResponse{
		CniReady:          true,
		ServiceReady:      true,
		PolicyReady:       true,
		GatewayReady:      true,
		KubeProxyReplaced: true,
		BpfMapRevision:    1,
		KernelVersion:     fmt.Sprintf("%d.%d.%d", kv.Major, kv.Minor, kv.Patch),
		NodeName:          s.nodeName,
		Version:           version.Version,
	}, nil
}

func main() {
	var (
		socketPath string
		bpfFSPath  string
		logLevel   string
		nodeName   string
		healthAddr string
		podCIDRStr string
	)

	flag.StringVar(&socketPath, "socket", defaultSocketPath,
		"Unix socket path for CNI and controller gRPC connections.")
	flag.StringVar(&bpfFSPath, "bpffs", defaultBPFFS,
		"BPF filesystem mount path for pinned maps and programs.")
	flag.StringVar(&logLevel, "log-level", "info",
		"Log level: debug, info, warn, error.")
	flag.StringVar(&nodeName, "node-name", os.Getenv("NODE_NAME"),
		"Kubernetes node name (injected via downward API).")
	flag.StringVar(&healthAddr, "health-addr", defaultHealthAddr,
		"Health probe HTTP bind address.")
	flag.StringVar(&podCIDRStr, "pod-cidr", os.Getenv("POD_CIDR"),
		"Node-local PodCIDR override (any valid CIDR, e.g. RFC 1918: 10.244.0.0/16, 172.16.0.0/16, 192.168.0.0/16; dynamically discovered if empty).")
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
	defer func() {
		if err := log.Sync(); err != nil {
			// Log sync errors but don't panic during shutdown
			_, _ = os.Stderr.WriteString("failed to sync logger: " + err.Error() + "\n")
		}
	}()

	// Admin CIDR override, dynamic node discovery, or RFC 1918 fallback.
	if podCIDRStr != "" {
		log.Info("using administrator-configured pod CIDR override", zap.String("podCIDR", podCIDRStr))
	} else if nodeName != "" {
		if k8sClient, err := api.NewClientset(); err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if discovered, err := api.DiscoverPodCIDR(ctx, k8sClient, nodeName); err == nil && discovered != "" {
				podCIDRStr = discovered
				log.Info("dynamically discovered node pod CIDR", zap.String("podCIDR", podCIDRStr))
			}
			cancel()
		}
	}

	// Fallback to RFC 1918 default pod CIDR if not specified or discovered.
	if podCIDRStr == "" {
		podCIDRStr = "10.244.0.0/16"
		log.Info("using default RFC 1918 pod CIDR", zap.String("podCIDR", podCIDRStr))
	}

	podCIDR, err := netip.ParsePrefix(podCIDRStr)
	if err != nil {
		log.Fatal("invalid pod-cidr configuration", zap.String("podCIDR", podCIDRStr), zap.Error(err))
	}

	log.Info("starting straitgatewayd node agent",
		zap.String("version", version.Version),
		zap.String("commit", version.Commit),
		zap.String("node", nodeName),
		zap.String("podCIDR", podCIDR.String()),
		zap.Bool("isRFC1918", podCIDR.Addr().IsPrivate()),
	)

	// Verify platform requirements (Linux kernel 6.6+ LTS / 6.7+).
	if err := process.CheckKernelVersion(log); err != nil {
		log.Warn("kernel version check warning", zap.Error(err))
	}
	if err := process.EnsureBPFFS(bpfFSPath); err != nil {
		log.Warn("BPF filesystem check warning", zap.Error(err))
	}

	// Initialize eBPF loader.
	bpfLoader := loader.New(log)
	if err := bpfLoader.LoadAll(); err != nil {
		log.Warn("loading eBPF programs (may require root or host bpffs)", zap.Error(err))
	}
	defer bpfLoader.Close() //nolint:errcheck

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Ensure socket directory exists.
	if err := os.MkdirAll("/run/straitgateway", 0750); err != nil {
		log.Fatal("creating socket directory", zap.Error(err))
	}

	// Remove old socket file, only fail on permission errors.
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		log.Fatal("removing old socket", zap.String("path", socketPath), zap.Error(err))
	}

	var lc net.ListenConfig
	lis, err := lc.Listen(ctx, "unix", socketPath)
	if err != nil {
		log.Fatal("listening on Unix socket", zap.String("path", socketPath), zap.Error(err))
	}
	defer os.Remove(socketPath)

	srv, err := newAgentServer(log, nodeName, bpfFSPath, podCIDR)
	if err != nil {
		log.Fatal("creating agent server", zap.Error(err))
	}

	grpcSrv := grpc.NewServer(
		grpc.MaxRecvMsgSize(4<<20),
		grpc.MaxSendMsgSize(4<<20),
	)
	agentv1.RegisterAgentServiceServer(grpcSrv, srv)

	// Error channel to propagate server failures to main.
	errChan := make(chan error, 2)

	// Start gRPC server.
	go func() {
		log.Info("agent gRPC server listening", zap.String("socket", socketPath))
		if err := grpcSrv.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			errChan <- err
		}
	}()

	// Start health HTTP server.
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	httpMux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready\n"))
	})
	healthSrv := &http.Server{
		Addr:    healthAddr,
		Handler: httpMux,
	}
	go func() {
		if err := healthSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or server error.
	select {
	case <-ctx.Done():
		log.Info("shutdown signal received, stopping straitgatewayd")
	case err := <-errChan:
		log.Error("server error, stopping straitgatewayd", zap.Error(err))
	}
	grpcSrv.GracefulStop()
	_ = healthSrv.Shutdown(context.Background())
	log.Info("straitgatewayd stopped successfully")
}
