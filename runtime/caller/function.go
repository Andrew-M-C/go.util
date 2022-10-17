package caller

import (
	"path/filepath"
	"strings"
)

// File identifies a full function path
type Function string

// Name returns the base function name.
func (f Function) Name() string {
	s := filepath.Ext(string(f))
	return strings.TrimLeft(s, ".")
}

// Package returns package name
func (f Function) Package() string {
	s := filepath.Base(string(f))
	parts := strings.Split(s, ".")
	return parts[0]
}

// Base returns the full base of a function.
func (f Function) Base() string {
	return filepath.Base(string(f))
}
