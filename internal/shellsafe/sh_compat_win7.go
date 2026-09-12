//go:build win7

package shellsafe

import "mvdan.cc/sh/v3/syntax"

// isClobberRedirectOp always reports false on the Win7 build. The Win7 build
// pins mvdan.cc/sh v3.7.0 (the newest release buildable under Go 1.20), which
// has no RdrClob/AppClob/RdrAllClob/AppAllClob operator constants because it
// predates parser support for the clobbering (>|) redirect forms. Such input
// simply never parses to those operators here, so treating them as absent is
// exact for this parser version.
func isClobberRedirectOp(op syntax.RedirOperator) bool {
	return false
}

// stmtDisownEffect always reports false on the Win7 build: sh v3.7.0 has no
// Disown field on Stmt (the `&|` disown operator postdates it).
func stmtDisownEffect(stmt *syntax.Stmt) bool {
	return false
}
