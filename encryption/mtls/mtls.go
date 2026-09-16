// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package mtls manages mutual TLS certificate loading and validation
// using Kubernetes CA and service account tokens.
package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"go.uber.org/zap"
)

// Config defines the paths to mTLS certificates and keys.
type Config struct {
	CACertPath     string
	ClientCertPath string
	ClientKeyPath  string
}

// Manager manages mTLS credentials for inter-cluster transit.
type Manager struct {
	log *zap.Logger
}

// New creates a new mTLS Manager.
func New(log *zap.Logger) *Manager {
	return &Manager{log: log}
}

// BuildClientTLSConfig constructs a tls.Config with the cluster CA and client certificate.
func (m *Manager) BuildClientTLSConfig(cfg Config) (*tls.Config, error) {
	caCert, err := os.ReadFile(cfg.CACertPath)
	if err != nil {
		return nil, fmt.Errorf("reading CA cert from %s: %w", cfg.CACertPath, err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA cert to pool")
	}

	cert, err := tls.LoadX509KeyPair(cfg.ClientCertPath, cfg.ClientKeyPath)
	if err != nil {
		return nil, fmt.Errorf("loading client key pair: %w", err)
	}

	return &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}, nil
}
