//go:build win7

package cli

import (
	"strings"
	"unicode/utf8"

	"reasonix/internal/compat"
)

// Plain-text styling for the reduced Win7 build. The full build's style.go and
// box.go render with the charm.land/lipgloss v2 stack and github.com/
// charmbracelet/x/ansi (Go >= 1.23), both excluded here. The non-interactive
// Win7 CLI emits unstyled text, so the color/emphasis helpers are identity
// functions and width measurement falls back to a rune count (no ANSI codes are
// ever produced, so there is nothing to strip).

func bold(s string) string    { return s }
func dim(s string) string     { return s }
func green(s string) string   { return s }
func accent(s string) string  { return s }
func reverse(s string) string { return s }
func red(s string) string     { return s }
func yellow(s string) string  { return s }

// ansiStrip is a no-op on Win7: this build emits no ANSI escapes, so external
// text needs no stripping (control-character filtering happens in the caller).
func ansiStrip(s string) string { return s }

// visibleWidth counts runes. Win7 output contains no ANSI SGR sequences, so a
// rune count is an adequate column estimate for the plain boxes below.
func visibleWidth(s string) int { return utf8.RuneCountInString(s) }

// boxed wraps content in a rounded box, mirroring the full build's layout but
// without color. The trailing newline lets callers Print it directly.
func boxed(lines []string) string {
	inner := 0
	for _, l := range lines {
		if w := visibleWidth(l); w > inner {
			inner = w
		}
	}
	inner += 2
	bar := strings.Repeat("─", inner)

	var b strings.Builder
	b.WriteString("╭" + bar + "╮")
	b.WriteByte('\n')
	for _, l := range lines {
		gap := compat.Max(inner-visibleWidth(l)-2, 0)
		b.WriteString("│ ")
		b.WriteString(l)
		b.WriteString(strings.Repeat(" ", gap))
		b.WriteString(" │")
		b.WriteByte('\n')
	}
	b.WriteString("╰" + bar + "╯")
	b.WriteByte('\n')
	return b.String()
}
