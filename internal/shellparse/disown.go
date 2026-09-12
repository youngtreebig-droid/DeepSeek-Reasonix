//go:build !win7

package shellparse

import "mvdan.cc/sh/v3/syntax"

// stmtDisown reports whether stmt uses the ksh `&|` disown operator. The field
// exists on the modern mvdan.cc/sh Stmt. The Win7 build pins the older sh
// v3.7.0 (the newest release buildable under Go 1.20), which predates the
// Disown field; see disown_win7.go.
func stmtDisown(stmt *syntax.Stmt) bool {
	return stmt.Disown
}
