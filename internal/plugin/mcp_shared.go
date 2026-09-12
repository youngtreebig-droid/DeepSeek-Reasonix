package plugin

import (
	"runtime/debug"
	"strings"
	"sync/atomic"
)

// This file holds the MCP session vocabulary shared by both the full build and
// the reduced Win7 build. The concrete MCP transport (built on
// github.com/modelcontextprotocol/go-sdk, which requires Go >= 1.23) lives in
// the `//go:build !win7` files sdk_session.go, sdk_errors.go, transport_http.go
// and friends. The Win7 build compiles none of the go-sdk; newTransport there
// returns a clear "MCP not available" error (see transport_win7.go). Server
// status reporting (server_status.go) still references these types on both
// builds, so they are defined here without any go-sdk dependency.

// SessionState is a coarse connection phase reported to /mcp status. It
// intentionally contains no endpoint, project path, or session identifier.
type SessionState string

const (
	SessionStateConnecting   SessionState = "connecting"
	SessionStateListening    SessionState = "listening"
	SessionStateReady        SessionState = "ready"
	SessionStateReconnecting SessionState = "reconnecting"
	SessionStateFailed       SessionState = "failed"
	SessionStateClosed       SessionState = "closed"
)

// SessionErrorKind classifies failures without exposing transport secrets.
type SessionErrorKind string

const (
	SessionErrorNone           SessionErrorKind = ""
	SessionErrorAuthRequired   SessionErrorKind = "auth_required"
	SessionErrorSessionMissing SessionErrorKind = "session_missing"
	SessionErrorStreamClosed   SessionErrorKind = "stream_closed"
	SessionErrorTimeout        SessionErrorKind = "timeout"
	SessionErrorProtocol       SessionErrorKind = "protocol"
	SessionErrorTransport      SessionErrorKind = "transport"
)

type sessionDiagnostics struct {
	ProtocolVersion   string
	State             SessionState
	SessionIDPresent  bool
	ReconnectAttempts int
	LastErrorKind     SessionErrorKind
	LastError         string
}

type sessionDiagnosticsProvider interface {
	sessionDiagnostics() sessionDiagnostics
}

var linkedMCPClientVersion atomic.Pointer[string]

// SetMCPClientVersion supplies the release version injected into an executable.
// Library and development builds fall back to module metadata or "dev".
func SetMCPClientVersion(version string) {
	version = strings.TrimSpace(version)
	if version == "" {
		version = "dev"
	}
	linkedMCPClientVersion.Store(&version)
}

func mcpClientVersion() string {
	if version := linkedMCPClientVersion.Load(); version != nil {
		return *version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
