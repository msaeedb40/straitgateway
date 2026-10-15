package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/straitgateway/straitgateway/internal/agent"
	"github.com/straitgateway/straitgateway/pkg/version"
)

type DaemonConfigFile struct {
	LogLevel            string `yaml:"logLevel"`
	BPFFsPath           string `yaml:"bpfFsPath"`
	CNIBinDir           string `yaml:"cniBinDir"`
	CNIConfDir          string `yaml:"cniConfDir"`
	ClusterCIDR         string `yaml:"clusterCIDR"`
	ServiceCIDR         string `yaml:"serviceCIDR"`
	TransitASN          uint32 `yaml:"transitASN"`
	WireGuardPort       int    `yaml:"wireguardPort"`
	WireGuardDev        string `yaml:"wireguardDev"`
	BFDEnabled          bool   `yaml:"bfdEnabled"`
	BFDMinRxInterval    string `yaml:"bfdMinRxInterval"`
	BFDMinTxInterval    string `yaml:"bfdMinTxInterval"`
	BFDDetectMultiplier int    `yaml:"bfdDetectMultiplier"`
	MetricsBindAddr     string `yaml:"metricsBindAddr"`
	NodeName            string `yaml:"nodeName"`
	NodeIP              string `yaml:"nodeIP"`
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthz" {
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("http://127.0.0.1:9090/healthz")
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}

	var configFile string
	var nodeName string
	var nodeIP string
	var localASN uint
	var metricsAddr string

	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.StringVar(&nodeName, "node-name", "", "Name of the Kubernetes node")
	flag.StringVar(&nodeIP, "node-ip", "", "Primary IP address of the node")
	flag.UintVar(&localASN, "asn", 0, "Local BGP Autonomous System Number")
	flag.StringVar(&metricsAddr, "metrics-bind-addr", "", "TCP address to bind Prometheus & HTTP observability server")
	flag.Parse()

	var fileCfg DaemonConfigFile
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to read config file %s: %v\n", configFile, err)
		} else {
			if err := yaml.Unmarshal(data, &fileCfg); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to parse config file %s: %v\n", configFile, err)
			}
		}
	}

	if nodeName == "" {
		if fileCfg.NodeName != "" {
			nodeName = fileCfg.NodeName
		} else if envNode := os.Getenv("NODE_NAME"); envNode != "" {
			nodeName = envNode
		} else {
			host, _ := os.Hostname()
			nodeName = host
		}
	}

	if nodeIP == "" {
		if fileCfg.NodeIP != "" {
			nodeIP = fileCfg.NodeIP
		} else if envIP := os.Getenv("NODE_IP"); envIP != "" {
			nodeIP = envIP
		} else {
			nodeIP = "127.0.0.1"
		}
	}

	if localASN == 0 {
		if fileCfg.TransitASN != 0 {
			localASN = uint(fileCfg.TransitASN)
		} else {
			localASN = 65001
		}
	}

	if metricsAddr == "" {
		if fileCfg.MetricsBindAddr != "" {
			metricsAddr = fileCfg.MetricsBindAddr
		} else {
			metricsAddr = ":9090"
		}
	}

	fmt.Printf("Starting straitd (StraitGateway Node Daemon) - %s\n", version.Info())

	cfg := agent.Config{
		NodeName:        nodeName,
		NodeIP:          net.ParseIP(nodeIP),
		LocalASN:        uint32(localASN),
		MetricsBindAddr: metricsAddr,
		BPFPinPath:      fileCfg.BPFFsPath,
		WireGuardPort:   fileCfg.WireGuardPort,
		WireGuardDev:    fileCfg.WireGuardDev,
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
