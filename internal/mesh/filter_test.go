package mesh

import (
	"net/http"
	"testing"
	"time"
)

func TestHeaderTransformFilter(t *testing.T) {
	filter := &HeaderTransformFilter{
		AddHeaders: map[string]string{
			"X-Custom-Env": "production",
			"X-Region":     "us-east-1",
		},
		RemoveHeaders: []string{"X-Deprecated-Header"},
	}

	headers := make(http.Header)
	headers.Set("X-Deprecated-Header", "old-value")

	ctx := &FilterContext{
		Headers:        headers,
		Attributes:     make(map[string]interface{}),
		ClientIdentity: 1001,
		Timestamp:      time.Now(),
	}

	res, err := filter.ProcessRequest(ctx)
	if err != nil {
		t.Fatalf("unexpected filter error: %v", err)
	}
	if res.Action != ActionTransform {
		t.Errorf("expected ActionTransform, got %s", res.Action)
	}
	if ctx.Headers.Get("X-Custom-Env") != "production" {
		t.Errorf("expected X-Custom-Env header added")
	}
	if ctx.Headers.Get("X-Deprecated-Header") != "" {
		t.Errorf("expected X-Deprecated-Header to be removed")
	}
}

func TestRateLimitFilter(t *testing.T) {
	rl := NewRateLimitFilter(2) // 2 requests per second

	ctx := &FilterContext{
		ClientIdentity: 1001,
		Attributes:     make(map[string]interface{}),
	}

	// Request 1: Allow
	res, _ := rl.ProcessRequest(ctx)
	if res.Action != ActionContinue {
		t.Errorf("req 1 should be allowed, got %s", res.Action)
	}

	// Request 2: Allow
	res, _ = rl.ProcessRequest(ctx)
	if res.Action != ActionContinue {
		t.Errorf("req 2 should be allowed, got %s", res.Action)
	}

	// Request 3: Abort with 429
	res, _ = rl.ProcessRequest(ctx)
	if res.Action != ActionAbort || res.StatusCode != http.StatusTooManyRequests {
		t.Errorf("req 3 should be 429 rate limited, got %+v", res)
	}
}

func TestJWTAuthFilter(t *testing.T) {
	jwtFilter := &JWTAuthFilter{ExpectedPrefix: "Bearer "}

	// 1. Missing header -> 401
	ctxNoHeader := &FilterContext{
		Headers:    make(http.Header),
		Attributes: make(map[string]interface{}),
	}
	res, _ := jwtFilter.ProcessRequest(ctxNoHeader)
	if res.Action != ActionAbort || res.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for missing auth header, got %+v", res)
	}

	// 2. Valid token -> Continue
	ctxValid := &FilterContext{
		Headers:    make(http.Header),
		Attributes: make(map[string]interface{}),
	}
	ctxValid.Headers.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiJ9.payload.signature")
	res, _ = jwtFilter.ProcessRequest(ctxValid)
	if res.Action != ActionContinue {
		t.Errorf("expected ActionContinue for valid bearer token, got %+v", res)
	}
	if ctxValid.Attributes["auth_token"] == nil {
		t.Errorf("expected auth_token attribute to be recorded")
	}
}

func TestWasmPluginLifecycleAndHotReload(t *testing.T) {
	wasmBinary := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00} // Valid Wasm magic header

	plugin, err := NewWasmPlugin(WasmPluginConfig{
		Name:     "geo-router",
		Bytecode: wasmBinary,
		Settings: map[string]string{
			"TargetCluster": "eu-central",
		},
	})
	if err != nil {
		t.Fatalf("failed to instantiate wasm plugin: %v", err)
	}

	ctx := &FilterContext{
		Headers:    make(http.Header),
		Attributes: make(map[string]interface{}),
	}

	res, err := plugin.ProcessRequest(ctx)
	if err != nil {
		t.Fatalf("failed to process request in wasm plugin: %v", err)
	}
	if res.Action != ActionTransform {
		t.Errorf("expected transform action, got %s", res.Action)
	}
	if res.Headers.Get("X-Wasm-TargetCluster") != "eu-central" {
		t.Errorf("expected X-Wasm-TargetCluster header to be eu-central")
	}

	// Hot reload with updated settings
	err = plugin.Reload(wasmBinary, map[string]string{
		"TargetCluster": "ap-southeast",
	})
	if err != nil {
		t.Fatalf("failed to reload wasm plugin: %v", err)
	}

	res, err = plugin.ProcessRequest(ctx)
	if err != nil {
		t.Fatalf("failed after reload: %v", err)
	}
	if res.Headers.Get("X-Wasm-TargetCluster") != "ap-southeast" {
		t.Errorf("expected updated setting ap-southeast, got %s", res.Headers.Get("X-Wasm-TargetCluster"))
	}
}
