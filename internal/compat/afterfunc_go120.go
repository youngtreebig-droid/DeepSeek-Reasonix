//go:build !go1.21

package compat

import "context"

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
