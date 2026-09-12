//go:build win7

package plugin

import (
	"context"
	"errors"
)

// ErrMCPUnavailableWin7 is returned whenever an MCP server would be started in
// the reduced Windows 7 build. The official MCP SDK
// (github.com/modelcontextprotocol/go-sdk) requires Go >= 1.23 and cannot be
// compiled by the go1.20.14 toolchain that produces Win7-runnable binaries, so
// the whole MCP transport is excluded from this build.
var ErrMCPUnavailableWin7 = errors.New("MCP servers are not available in the Win7 build")

// newTransport reports that MCP is disabled on the Win7 build. Callers surface
// this as a per-server startup failure; the rest of the CLI keeps working
// without any configured MCP servers.
func newTransport(ctx context.Context, s Spec, profile HostProfile) (transport, error) {
	return nil, ErrMCPUnavailableWin7
}
