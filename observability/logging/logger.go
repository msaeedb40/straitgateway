// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package logging provides structured JSON logging via go.uber.org/zap.
// All straitgateway components use this logger factory.
package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new production zap.Logger with ISO8601 timestamps.
func New(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.StacktraceKey = ""

	switch level {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	return cfg.Build(zap.AddCallerSkip(0), zap.AddCaller())
}

// NewDevelopment creates a development logger with human-readable output.
func NewDevelopment() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return cfg.Build()
}
