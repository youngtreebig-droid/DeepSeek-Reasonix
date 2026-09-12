//go:build win7

package cli

import (
	"fmt"
	"os"
)

// defaultInteractiveSession reports that the interactive TUI is unavailable on
// the reduced Win7 build. The charm.land/v2 TUI stack (bubbletea/bubbles/
// lipgloss) requires Go >= 1.23 and is excluded from the go1.20.14 build; use
// the non-interactive subcommands (run, serve, config, ...) instead.
func defaultInteractiveSession(args []string, version string) int {
	fmt.Fprintln(os.Stderr, "reasonix: the interactive TUI is not available in the Win7 build; use non-interactive commands such as `reasonix run --print \"...\"` or `reasonix serve`.")
	return 2
}
