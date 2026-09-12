//go:build windows && win7

package notify

// Win7 build note: the full Windows sender uses git.sr.ht/~jackmordaunt/go-toast/v2
// for toast notifications, but that package's wintoast implementation uses
// runtime.Pinner (Go 1.21), which the go1.20.14 toolchain that targets Windows 7
// cannot compile. The reduced Win7 build therefore ships a no-op desktop
// notifier: the CLI still runs, it just does not raise Windows toast popups.

// PlatformSender is the no-op Windows notifier for the reduced Win7 build.
type PlatformSender struct{}

// NewPlatformSender returns the best-effort sender for the current platform.
func NewPlatformSender() PlatformSender { return PlatformSender{} }

// Send discards the notification on the Win7 build; toast delivery is
// unavailable without the go-toast dependency.
func (PlatformSender) Send(Message) error { return nil }
