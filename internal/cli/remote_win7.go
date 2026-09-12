//go:build win7

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// remoteCommand is stubbed on the reduced Win7 build. The SSH remote subsystem
// (internal/remote) depends on golang.org/x/crypto/ssh's client AuthCallback
// API, which only exists in x/crypto releases that require Go >= 1.25 and so
// cannot compile under the go1.20.14 Win7 toolchain. Use the full build for
// `reasonix remote`.
func remoteCommand(args []string, version string) int {
	fmt.Fprintln(os.Stderr, "reasonix: `remote` (SSH remote workspaces) is not available in the Win7 build; use the full build on a supported OS.")
	return 2
}

// openInBrowser launches the platform browser for url. This is a plain
// exec-based helper with no non-go-1.20 dependencies, so the Win7 build keeps
// the real implementation rather than a stub (used by `reasonix web`).
func openInBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{url}
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default:
		cmd, args = "xdg-open", []string{url}
	}
	c := exec.Command(cmd, args...)
	if err := c.Start(); err != nil {
		return err
	}
	go func() { _ = c.Wait() }()
	return nil
}
