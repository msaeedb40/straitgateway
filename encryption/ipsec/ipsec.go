// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package ipsec manages Linux XFRM IPsec tunnels and security associations
// for transit encryption between clusters.
package ipsec

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

// PolicyConfig defines an XFRM IPsec policy configuration.
type PolicyConfig struct {
	SrcNet *net.IPNet
	DstNet *net.IPNet
	SPI    int
	ReqID  int
}

// Manager manages Linux IPsec XFRM states and policies.
type Manager struct {
	log *zap.Logger
}

// New creates a new IPsec Manager.
func New(log *zap.Logger) *Manager {
	return &Manager{log: log}
}

// AddXFRMPolicy adds an XFRM policy for transit encryption.
func (m *Manager) AddXFRMPolicy(cfg PolicyConfig) error {
	policy := &netlink.XfrmPolicy{
		Src: cfg.SrcNet,
		Dst: cfg.DstNet,
		Dir: netlink.XFRM_DIR_OUT,
		Tmpls: []netlink.XfrmPolicyTmpl{
			{
				Src:   cfg.SrcNet.IP,
				Dst:   cfg.DstNet.IP,
				Proto: netlink.XFRM_PROTO_ESP,
				Mode:  netlink.XFRM_MODE_TUNNEL,
				Reqid: cfg.ReqID,
			},
		},
	}

	if err := netlink.XfrmPolicyAdd(policy); err != nil {
		return fmt.Errorf("adding XFRM policy (%s -> %s): %w", cfg.SrcNet, cfg.DstNet, err)
	}

	m.log.Info("added XFRM IPsec policy",
		zap.String("src", cfg.SrcNet.String()),
		zap.String("dst", cfg.DstNet.String()),
		zap.Int("spi", cfg.SPI),
	)
	return nil
}

// DeleteXFRMPolicy removes an XFRM policy.
func (m *Manager) DeleteXFRMPolicy(cfg PolicyConfig) error {
	policy := &netlink.XfrmPolicy{
		Src: cfg.SrcNet,
		Dst: cfg.DstNet,
		Dir: netlink.XFRM_DIR_OUT,
	}
	return netlink.XfrmPolicyDel(policy)
}
