package ktls

import (
	"testing"
)

func TestKTLSManagerLifecycle(t *testing.T) {
	mgr := NewManager()

	if !mgr.IsSupported() {
		t.Fatal("expected kTLS to be supported in mock/kernel mode")
	}

	tx := &CryptoInfo{
		Cipher:  CipherAESGCM128,
		Version: TLSVersion13,
		Key:     []byte("0123456789abcdef"),
		IV:      []byte("1234"),
	}
	rx := &CryptoInfo{
		Cipher:  CipherAESGCM128,
		Version: TLSVersion13,
		Key:     []byte("fedcba9876543210"),
		IV:      []byte("4321"),
	}

	// 1. Configure socket kTLS
	session, err := mgr.ConfigureSocketKTLS("sess-01", 10, tx, rx, true)
	if err != nil {
		t.Fatalf("failed to configure socket kTLS: %v", err)
	}
	if session.SocketFD != 10 || session.Cipher != CipherAESGCM128 {
		t.Errorf("unexpected session parameters: %+v", session)
	}
	if mgr.ActiveSessionsCount() != 1 {
		t.Errorf("expected 1 active session, got %d", mgr.ActiveSessionsCount())
	}

	// 2. Record encrypted transfers
	err = mgr.RecordTransfer("sess-01", 1024, 2048)
	if err != nil {
		t.Fatalf("failed to record transfer: %v", err)
	}
	s, exists := mgr.GetSession("sess-01")
	if !exists || s.BytesEncrypted != 1024 || s.BytesDecrypted != 2048 {
		t.Errorf("expected 1024 tx and 2048 rx bytes, got %+v", s)
	}

	// 3. Close session
	mgr.CloseSession("sess-01")
	if mgr.ActiveSessionsCount() != 0 {
		t.Errorf("expected 0 active sessions after close")
	}
}

func TestKTLSValidationErrors(t *testing.T) {
	mgr := NewManager()

	// Invalid FD
	_, err := mgr.ConfigureSocketKTLS("sess-bad", 0, &CryptoInfo{Key: []byte("k")}, &CryptoInfo{Key: []byte("k")}, false)
	if err == nil {
		t.Error("expected error for FD <= 0")
	}

	// Empty keys
	_, err = mgr.ConfigureSocketKTLS("sess-bad", 5, &CryptoInfo{}, &CryptoInfo{}, false)
	if err == nil {
		t.Error("expected error for empty crypto keys")
	}
}
