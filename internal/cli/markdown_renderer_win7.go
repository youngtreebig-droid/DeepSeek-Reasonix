//go:build win7

package cli

import "reasonix/internal/agent"

// cliMarkdownRenderer returns nil on the reduced Win7 build so `run` emits plain
// text. The ANSI markdown renderer (md.go) is built on the charm.land/x-ansi
// styling stack, which requires Go >= 1.23 and is excluded from the go1.20.14
// build. agent.NewTextSink treats a nil renderer as pass-through text.
func cliMarkdownRenderer(width int) agent.Renderer {
	return nil
}
