//go:build go1.21

package compat

import "context"

// ContextAfterFunc forwards to context.AfterFunc (added in Go 1.21). On every
// toolchain that has it, this is the exact standard behaviour, so the default
// build (and any go1.21+ build) uses the std implementation with no extra
// goroutine bookkeeping. The go1.20-only variant lives in afterfunc_go120.go.
func ContextAfterFunc(ctx context.Context, f func()) (stop func() bool) {
	return context.AfterFunc(ctx, f)
}
