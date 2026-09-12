//go:build !win7

package shellsafe

import "mvdan.cc/sh/v3/syntax"

// isClobberRedirectOp reports whether op is one of the clobbering write
// redirect operators (>|, &>|, and their all-streams forms). Those operator
// constants exist on the modern mvdan.cc/sh syntax package. The Win7 build
// pins the older sh v3.7.0, which predates them; see sh_compat_win7.go.
func isClobberRedirectOp(op syntax.RedirOperator) bool {
	switch op {
	case syntax.RdrClob, syntax.AppClob, syntax.RdrAllClob, syntax.AppAllClob:
		return true
	default:
		return false
	}
}

// stmtDisownEffect reports whether stmt uses the ksh `&|` disown operator.
func stmtDisownEffect(stmt *syntax.Stmt) bool {
	return stmt.Disown
}
