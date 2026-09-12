//go:build win7

package cli

// Minimal theme surface for the reduced Win7 build. The full CLI theme
// (theme.go) is built on the charm.land/lipgloss v2 styling stack, which
// requires Go >= 1.23 and is excluded from the go1.20.14 build along with the
// interactive TUI. The non-interactive Win7 CLI emits plain, unstyled text, so
// these are no-op stubs covering only the symbols the core router still calls.

// terminalRGB mirrors the full build's background-probe color; the Win7 build
// never probes the terminal, so only the empty value is ever produced.
type terminalRGB struct {
	r int
	g int
	b int
}

// cliCursorShape is unused without the interactive composer but is read by the
// config wiring in cli.go; keep the default so that code path compiles.
var cliCursorShape = "bar"

// configureCLITheme and configureCLIThemeWithStyle are no-ops on Win7: there is
// no lipgloss palette to resolve. Output is emitted unstyled.
func configureCLITheme(mode string)                {}
func configureCLIThemeWithStyle(mode, style string) {}

// withTerminalProbe runs fn directly. The full build first probes the terminal
// background over OSC 11 to pick a light/dark palette; the Win7 build has no
// palette to pick, so it just invokes fn.
func withTerminalProbe(fn func()) { fn() }
