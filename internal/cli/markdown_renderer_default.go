//go:build !win7

package cli

import (
	"github.com/charmbracelet/x/ansi"

	"reasonix/internal/agent"
)

// cliMarkdownRenderer returns the ANSI markdown renderer for TTY `run` output.
// The Win7 build returns nil (plain text) because the renderer depends on the
// charm.land/x-ansi styling stack (Go >= 1.23); see markdown_renderer_win7.go.
func cliMarkdownRenderer(width int) agent.Renderer {
	return newMarkdownRenderer(width)
}

// ansiStrip removes ANSI escape sequences from s. The Win7 build (style_win7.go)
// uses a no-op because it never emits escapes.
func ansiStrip(s string) string {
	return ansi.Strip(s)
}
