// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package systemd provides service supervision integration with systemd.
package systemd

import (
	"fmt"
	"net"
	"os"

	"go.uber.org/zap"
)

// Notifier notifies systemd of service status via sd_notify protocol.
type Notifier struct {
	socketPath string
	log        *zap.Logger
}

// NewNotifier creates a new systemd Notifier reading NOTIFY_SOCKET.
func NewNotifier(log *zap.Logger) *Notifier {
	return &Notifier{
		socketPath: os.Getenv("NOTIFY_SOCKET"),
		log:        log,
	}
}

// NotifyReady signals systemd that the service is ready (READY=1).
func (n *Notifier) NotifyReady() error {
	return n.send("READY=1")
}

// NotifyWatchdog sends a heartbeat ping to systemd (WATCHDOG=1).
func (n *Notifier) NotifyWatchdog() error {
	return n.send("WATCHDOG=1")
}

// NotifyStatus sends a descriptive status message to systemd (STATUS=...).
func (n *Notifier) NotifyStatus(status string) error {
	return n.send(fmt.Sprintf("STATUS=%s", status))
}

func (n *Notifier) send(state string) error {
	if n.socketPath == "" {
		return nil // Not running under systemd with notification socket
	}

	conn, err := net.Dial("unixgram", n.socketPath)
	if err != nil {
		return fmt.Errorf("dialing NOTIFY_SOCKET: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(state))
	return err
}
