//go:build go1.21

package compat

import (
	"context"
	"time"
)

// ContextWithTimeoutCause forwards to context.WithTimeoutCause (added in
// Go 1.21). The go1.20 variant drops the cause (see afterfunc_go120.go).
func ContextWithTimeoutCause(parent context.Context, timeout time.Duration, cause error) (context.Context, context.CancelFunc) {
	return context.WithTimeoutCause(parent, timeout, cause)
}

// ContextWithoutCancel forwards to context.WithoutCancel (added in Go 1.21):
// it returns a copy of parent that is not cancelled when parent is.
func ContextWithoutCancel(parent context.Context) context.Context {
	return context.WithoutCancel(parent)
}

// ContextAfterFunc forwards to context.AfterFunc (added in Go 1.21). On every
// toolchain that has it, this is the exact standard behaviour, so the default
// build (and any go1.21+ build) uses the std implementation with no extra
// goroutine bookkeeping. The go1.20-only variant lives in afterfunc_go120.go.
func ContextAfterFunc(ctx context.Context, f func()) (stop func() bool) {
	return context.AfterFunc(ctx, f)
}
