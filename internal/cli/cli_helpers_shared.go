package cli

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"unicode"

	"reasonix/internal/config"
	"reasonix/internal/gitcmd"
)

// These small helpers are plain logic with no interactive-TUI dependency, so
// they live in an untagged file shared by both the full build and the reduced
// Win7 build. They were previously colocated with TUI-only code that the Win7
// build excludes.

// mcpOpenCommand builds the platform command that opens target (a URL or file)
// in the user's default handler. Used by the MCP OAuth browser step.
func mcpOpenCommand(target string) (*exec.Cmd, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil, fmt.Errorf("empty target")
	}
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", target), nil
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", target), nil
	default:
		return exec.Command("xdg-open", target), nil
	}
}

// hasUnescapedPathWhitespace reports whether s contains whitespace that is not
// backslash-escaped, used to decide when a pasted path needs quoting.
func hasUnescapedPathWhitespace(s string) bool {
	escaped := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			return true
		}
	}
	return false
}

// unescapeShellPath removes one level of backslash escaping from a pasted path.
func unescapeShellPath(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// modelRefs lists the "provider/model" references from configured providers.
func modelRefs() []string {
	cfg, err := config.Load()
	if err != nil {
		return nil
	}
	var out []string
	for i := range cfg.Providers {
		p := &cfg.Providers[i]
		if !p.Configured() {
			continue
		}
		for _, model := range p.ChatModelList() {
			out = append(out, p.Name+"/"+model)
		}
	}
	return out
}

// runGit runs a git command in cwd and returns its stdout. cwd flows through
// gitcmd's dir parameter so the gitcmd baseline resolves the repository config
// relative to it.
func runGit(ctx context.Context, cwd string, args ...string) (string, error) {
	cmd := gitcmd.Command(ctx, cwd, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// oneLineText collapses external display text to a single sanitized line.
func oneLineText(s string) string {
	return sanitizeExternalDisplayText(s)
}

// sanitizeExternalDisplayText strips ANSI escapes and control characters from
// externally sourced text and collapses runs of whitespace to single spaces, so
// untrusted content cannot inject escape sequences or line breaks into CLI
// output. ANSI stripping is build-tagged (ansiStrip) because the full build uses
// x/ansi while the Win7 build, whose own output carries no escapes, uses a
// no-op.
func sanitizeExternalDisplayText(s string) string {
	s = ansiStrip(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\t' || r == '\n' || r == '\r':
			b.WriteByte(' ')
		case r < 0x20 || r == 0x7f:
			// Drop remaining C0 controls and DEL.
		case r >= 0x80 && r <= 0x9f:
			// Drop C1 controls (including after partial decode).
		case unicode.Is(unicode.Cc, r):
			// Other control categories.
		default:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
