//go:build win7

package shellparse

import "mvdan.cc/sh/v3/syntax"

// stmtDisown always reports false on the Win7 build. The Win7 build pins
// mvdan.cc/sh v3.7.0 (the newest release still buildable under Go 1.20), which
// has no Disown field on Stmt because it predates parser support for the ksh
// `&|` disown operator. Such input parses without the disown flag, so the
// static command policy simply never sees a disowned statement here.
func stmtDisown(stmt *syntax.Stmt) bool {
	return false
}
