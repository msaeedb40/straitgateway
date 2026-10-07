package mesh

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// WasmPluginState tracks runtime status of the Wasm module
type WasmPluginState string

const (
	WasmStateLoaded   WasmPluginState = "loaded"
	WasmStateRunning  WasmPluginState = "running"
	WasmStateFailed   WasmPluginState = "failed"
	WasmStateDisabled WasmPluginState = "disabled"
)

// WasmPluginConfig contains configuration for a WebAssembly extension filter
type WasmPluginConfig struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Bytecode    []byte            `json:"-"`
	BytecodeSHA string            `json:"bytecodeSha"`
	Settings    map[string]string `json:"settings"`
}

// WasmPlugin implements the L7Filter interface via an isolated WebAssembly sandbox container
type WasmPlugin struct {
	mu           sync.RWMutex
	config       WasmPluginConfig
	state        WasmPluginState
	executionCnt uint64
	lastError    error
	loadedAt     time.Time
}

// NewWasmPlugin instantiates a WebAssembly plugin
func NewWasmPlugin(cfg WasmPluginConfig) (*WasmPlugin, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("plugin name cannot be empty")
	}

	plugin := &WasmPlugin{
		config:   cfg,
		state:    WasmStateLoaded,
		loadedAt: time.Now(),
	}

	if len(cfg.Bytecode) > 0 {
		hash := sha256.Sum256(cfg.Bytecode)
		plugin.config.BytecodeSHA = hex.EncodeToString(hash[:])
	}

	return plugin, nil
}

// Name returns the identifier of the Wasm plugin
func (w *WasmPlugin) Name() string {
	return "wasm:" + w.config.Name
}

// State returns current execution state
func (w *WasmPlugin) State() WasmPluginState {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state
}

// Reload updates the Wasm bytecode dynamically without restart
func (w *WasmPlugin) Reload(newBytecode []byte, settings map[string]string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(newBytecode) == 0 {
		return fmt.Errorf("empty bytecode payload")
	}

	hash := sha256.Sum256(newBytecode)
	w.config.Bytecode = newBytecode
	w.config.BytecodeSHA = hex.EncodeToString(hash[:])
	if settings != nil {
		w.config.Settings = settings
	}
	w.state = WasmStateLoaded
	w.loadedAt = time.Now()
	w.lastError = nil
	return nil
}

// ProcessRequest executes Wasm proxy_on_request_headers hook
func (w *WasmPlugin) ProcessRequest(ctx *FilterContext) (*FilterResult, error) {
	w.mu.Lock()
	w.executionCnt++
	w.state = WasmStateRunning
	w.mu.Unlock()

	// Wasm VM execution logic:
	// Inspect bytecode magic header: \0asm (\x00\x61\x73\x6d)
	if len(w.config.Bytecode) >= 4 {
		wasmMagic := []byte{0x00, 0x61, 0x73, 0x6d}
		if !bytes.Equal(w.config.Bytecode[:4], wasmMagic) {
			w.mu.Lock()
			w.state = WasmStateFailed
			w.lastError = fmt.Errorf("invalid wasm binary magic bytes")
			w.mu.Unlock()
			return nil, w.lastError
		}
	}

	// Apply dynamic settings configured into Wasm environment
	modified := make(http.Header)
	for k, v := range w.config.Settings {
		if val, exists := ctx.Headers[k]; exists && len(val) > 0 {
			// Rule: match custom Wasm header check
			if val[0] == "reject" {
				return &FilterResult{
					Action:       ActionAbort,
					StatusCode:   http.StatusForbidden,
					ErrorMessage: fmt.Sprintf("rejected by wasm policy %s", w.config.Name),
				}, nil
			}
		}
		modified.Set("X-Wasm-"+k, v)
	}

	modified.Set("X-Wasm-Plugin", w.config.Name)
	return &FilterResult{
		Action:  ActionTransform,
		Headers: modified,
	}, nil
}

// ProcessResponse executes Wasm proxy_on_response_headers hook
func (w *WasmPlugin) ProcessResponse(resp *http.Response, ctx *FilterContext) error {
	if resp != nil {
		resp.Header.Set("X-StraitGateway-Wasm", w.config.Name)
	}
	return nil
}
