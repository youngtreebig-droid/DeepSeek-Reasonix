//go:build !go1.21

package compat

import (
	"context"
	"time"
)

// ContextWithTimeoutCause backports context.WithTimeoutCause for the go1.20.14
// Win7 toolchain. Go 1.20 has no cause-carrying variant, so the cause is
// dropped and a plain deadline context is returned; ctx.Err() still reports
// context.DeadlineExceeded on timeout, only context.Cause(ctx) differs.
func ContextWithTimeoutCause(parent context.Context, timeout time.Duration, cause error) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}

// withoutCancel wraps a parent context but is never cancelled and has no
// deadline. It preserves value lookups (Value) so request-scoped data still
// flows through, matching context.WithoutCancel (added in Go 1.21).
type withoutCancel struct{ parent context.Context }

func (withoutCancel) Deadline() (time.Time, bool) { return time.Time{}, false }
func (withoutCancel) Done() <-chan struct{}       { return nil }
func (withoutCancel) Err() error                  { return nil }
func (c withoutCancel) Value(key any) any         { return c.parent.Value(key) }

// ContextWithoutCancel backports context.WithoutCancel for Go 1.20: it returns
// a copy of parent that is not cancelled when parent is, retaining values.
func ContextWithoutCancel(parent context.Context) context.Context {
	return withoutCancel{parent: parent}
}

// ContextAfterFunc backports context.AfterFunc for the go1.20.14 Win7
// toolchain, which predates it. It arranges to call f in its own goroutine
// after ctx is done (cancelled or its deadline expires) and returns a stop
// function that unregisters f; stop returns true if it prevented f from
// running. The watcher goroutine always exits once ctx is done or stop is
// called, matching the standard function's lifecycle.
func ContextAfterFunc(ctx context.Context, f func()) (stop func() bool) {
	stopc := make(chan struct{})
	stopped := make(chan bool, 1)
	go func() {
		select {
		case <-ctx.Done():
			go f()
			stopped <- false
		case <-stopc:
			stopped <- true
		}
	}()
	var once bool
	return func() bool {
		if once {
			return false
		}
		once = true
		close(stopc)
		return <-stopped
	}
}
