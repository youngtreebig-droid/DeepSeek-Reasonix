//go:build !win7

package plugin

import "context"

// newTransport builds the transport for a spec's declared type. Empty / unknown
// defaults to stdio. The concrete transport is backed by the official MCP SDK
// (github.com/modelcontextprotocol/go-sdk).
func newTransport(ctx context.Context, s Spec, profile HostProfile) (transport, error) {
	return newSDKSessionTransport(ctx, s, profile)
}
