// Package serveenv holds the ssh-free constants and serve-state record shared
// between the SSH bootstrap layer (internal/remote/bootstrap) and the local
// `reasonix serve`/session CLI. It intentionally imports nothing from
// internal/remote or golang.org/x/crypto/ssh so the non-remote CLI paths that
// only need these values can compile on the go1.20.14 Win7 toolchain, where
// the ssh client stack is excluded.
package serveenv

import "encoding/json"

// Environment variables and capability markers exchanged with a bootstrapped
// serve process. These are plain strings with no ssh dependency.
const (
	// BrowserBrokerEnv carries the desktop browser broker base URL to a serve
	// child.
	BrowserBrokerEnv = "REASONIX_BROWSER_BROKER"
	// BrowserTokenEnv carries the desktop browser broker token to a serve child.
	BrowserTokenEnv = "REASONIX_BROWSER_TOKEN"
	// ServeBrowserBrokerMarker is the serve flag/marker enabling the browser
	// broker.
	ServeBrowserBrokerMarker = "browser-broker"
	// ServeCapsToken marks a serve build that supports the required serve
	// capabilities without a runtime probe.
	ServeCapsToken = "reasonix-serve-caps-20260826a"
)

// ServeState is the JSON record a bootstrapped serve leaves so a later
// (re)connect can find and reuse it. Fields use omitempty so an older record
// missing a field still decodes.
type ServeState struct {
	PID       int    `json:"pid"`
	Addr      string `json:"addr"` // 127.0.0.1:<port> on the remote host
	Workspace string `json:"workspace"`
	Version   string `json:"version,omitempty"`
	ServeCaps string `json:"serve_caps,omitempty"`
	TokenFile string `json:"token_file"`
	LogFile   string `json:"log_file,omitempty"`
	StartedAt int64  `json:"started_at,omitempty"` // unix seconds
}

// MarshalState renders a ServeState as indented JSON.
func MarshalState(s ServeState) ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// UnmarshalState parses a ServeState record.
func UnmarshalState(data []byte) (ServeState, error) {
	var s ServeState
	if err := json.Unmarshal(data, &s); err != nil {
		return ServeState{}, err
	}
	return s, nil
}
