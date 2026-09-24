// Package config holds StraitGateway runtime configuration.
package config

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// Validate checks the Config struct for correctness and consistent runtime settings.
func (c *Config) Validate() error {
	var errs []string

	// Datapath validation
	if err := validateDatapath(&c.Datapath); err != nil {
		errs = append(errs, fmt.Sprintf("datapath: %v", err))
	}

	// CNI validation
	if err := validateCNI(&c.CNI); err != nil {
		errs = append(errs, fmt.Sprintf("cni: %v", err))
	}

	// IPAM validation
	if err := validateIPAM(&c.IPAM); err != nil {
		errs = append(errs, fmt.Sprintf("ipam: %v", err))
	}

	// Transit validation
	if c.Transit.Enabled {
		if err := validateTransit(&c.Transit); err != nil {
			errs = append(errs, fmt.Sprintf("transit: %v", err))
		}
	}

	// Observability validation
	if err := validateObservability(&c.Observability); err != nil {
		errs = append(errs, fmt.Sprintf("observability: %v", err))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func validateDatapath(d *DatapathConfig) error {
	mode := strings.ToLower(d.Mode)
	if mode != "" && mode != "netkit" && mode != "veth" {
		return fmt.Errorf("invalid datapath mode %q; must be 'netkit' or 'veth'", d.Mode)
	}
	return nil
}

func validateCNI(c *CNIConfig) error {
	if c.MTU < 0 || c.MTU > 65535 {
		return fmt.Errorf("invalid MTU %d; must be between 0 and 65535", c.MTU)
	}
	if c.LogLevel != "" && !isValidLogLevel(c.LogLevel) {
		return fmt.Errorf("invalid log level %q", c.LogLevel)
	}
	return nil
}

func validateIPAM(i *IPAMConfig) error {
	mode := strings.ToLower(i.Mode)
	if mode != "" && mode != "cluster-pool" && mode != "multi-pool" {
		return fmt.Errorf("invalid IPAM mode %q; must be 'cluster-pool' or 'multi-pool'", i.Mode)
	}
	for _, cidr := range i.PodCIDRs {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return fmt.Errorf("invalid pod CIDR %q: %w", cidr, err)
		}
	}
	return nil
}

func validateTransit(t *TransitConfig) error {
	topology := strings.ToLower(t.Topology)
	switch topology {
	case "hub-spoke", "mesh", "peer-to-peer", "hub-to-hub", "hybrid":
		return nil
	default:
		return fmt.Errorf("invalid topology %q; must be hub-spoke, mesh, peer-to-peer, hub-to-hub, or hybrid", t.Topology)
	}
}

func validateObservability(o *ObservabilityConfig) error {
	if o.LogLevel != "" && !isValidLogLevel(o.LogLevel) {
		return fmt.Errorf("invalid log level %q", o.LogLevel)
	}
	if o.LogFormat != "" {
		fmtLower := strings.ToLower(o.LogFormat)
		if fmtLower != "json" && fmtLower != "text" {
			return fmt.Errorf("invalid log format %q; must be 'json' or 'text'", o.LogFormat)
		}
	}
	return nil
}

func isValidLogLevel(lvl string) bool {
	switch strings.ToLower(lvl) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}
