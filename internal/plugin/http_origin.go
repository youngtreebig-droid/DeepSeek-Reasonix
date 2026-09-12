package plugin

import (
	"net/url"
	"strings"
)

// sameHTTPOrigin reports whether a and b share scheme, host, and effective
// port. It is used by both the MCP HTTP transport (default build) and OAuth
// origin checks, so it lives in a build-tag-neutral file.
func sameHTTPOrigin(a, b *url.URL) bool {
	if a == nil || b == nil || !strings.EqualFold(a.Scheme, b.Scheme) || !strings.EqualFold(a.Hostname(), b.Hostname()) {
		return false
	}
	effectivePort := func(u *url.URL) string {
		if port := u.Port(); port != "" {
			return port
		}
		switch strings.ToLower(u.Scheme) {
		case "http":
			return "80"
		case "https":
			return "443"
		default:
			return ""
		}
	}
	return effectivePort(a) == effectivePort(b)
}
