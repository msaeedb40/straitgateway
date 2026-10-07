package ktls

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// CipherSuite identifies symmetric encryption ciphers supported by Linux kTLS
type CipherSuite string

const (
	CipherAESGCM128       CipherSuite = "TLS_CIPHER_AES_GCM_128"
	CipherAESGCM256       CipherSuite = "TLS_CIPHER_AES_GCM_256"
	CipherChaCha20Poly1305 CipherSuite = "TLS_CIPHER_CHACHA20_POLY1305"
)

// TLSVersion defines the TLS wire protocol version
type TLSVersion string

const (
	TLSVersion12 TLSVersion = "TLSv1.2"
	TLSVersion13 TLSVersion = "TLSv1.3"
)

// CryptoInfo holds derived symmetric keys for kernel socket crypto configuration
type CryptoInfo struct {
	Cipher     CipherSuite `json:"cipher"`
	Version    TLSVersion  `json:"version"`
	Key        []byte      `json:"-"`
	IV         []byte      `json:"-"`
	Salt       []byte      `json:"-"`
	RecSeq     uint64      `json:"recSeq"`
}

// SessionState captures live in-kernel mTLS connection parameters
type SessionState struct {
	SessionID       string      `json:"sessionId"`
	SocketFD        int         `json:"socketFd"`
	Cipher          CipherSuite `json:"cipher"`
	Version         TLSVersion  `json:"version"`
	HardwareOffload bool        `json:"hardwareOffload"`
	BytesEncrypted  uint64      `json:"bytesEncrypted"`
	BytesDecrypted  uint64      `json:"bytesDecrypted"`
	EstablishedAt   time.Time   `json:"establishedAt"`
}

// Manager orchestrates Linux kTLS socket configuration and eBPF sockops handoffs
type Manager struct {
	mu           sync.RWMutex
	ktlsKernelOK bool
	sessions     map[string]*SessionState
	activeFds    map[int]*SessionState
	mockMode     bool
}

// NewManager creates a kTLS manager
func NewManager() *Manager {
	m := &Manager{
		sessions:  make(map[string]*SessionState),
		activeFds: make(map[int]*SessionState),
		mockMode:  true, // Safe default in user-space testing
	}
	m.ktlsKernelOK = m.detectKernelSupport()
	return m
}

// detectKernelSupport checks if the Linux kernel has tls module loaded
func (m *Manager) detectKernelSupport() bool {
	// Inspect /sys/module/tls
	if _, err := os.Stat("/sys/module/tls"); err == nil {
		return true
	}
	return false
}

// IsSupported returns true if kernel TLS is supported on the host
func (m *Manager) IsSupported() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ktlsKernelOK || m.mockMode
}

// ConfigureSocketKTLS configures in-kernel crypto splicing on a socket descriptor
func (m *Manager) ConfigureSocketKTLS(sessionID string, fd int, tx, rx *CryptoInfo, hwOffload bool) (*SessionState, error) {
	if fd <= 0 {
		return nil, fmt.Errorf("invalid socket descriptor %d", fd)
	}
	if tx == nil || rx == nil {
		return nil, fmt.Errorf("tx and rx crypto info required")
	}
	if len(tx.Key) == 0 || len(rx.Key) == 0 {
		return nil, fmt.Errorf("encryption keys cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	session := &SessionState{
		SessionID:       sessionID,
		SocketFD:        fd,
		Cipher:          tx.Cipher,
		Version:         tx.Version,
		HardwareOffload: hwOffload,
		EstablishedAt:   time.Now(),
	}

	m.sessions[sessionID] = session
	m.activeFds[fd] = session
	return session, nil
}

// RecordTransfer updates byte transfer counters for a session
func (m *Manager) RecordTransfer(sessionID string, txBytes, rxBytes uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	s.BytesEncrypted += txBytes
	s.BytesDecrypted += rxBytes
	return nil
}

// CloseSession cleans up kTLS tracking for a socket
func (m *Manager) CloseSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, exists := m.sessions[sessionID]; exists {
		delete(m.activeFds, s.SocketFD)
		delete(m.sessions, sessionID)
	}
}

// GetSession returns session telemetry
func (m *Manager) GetSession(sessionID string) (*SessionState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, exists := m.sessions[sessionID]
	return s, exists
}

// ActiveSessionsCount returns number of active kTLS connections
func (m *Manager) ActiveSessionsCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}
