//go:build win7

package agent

import "mvdan.cc/sh/v3/syntax"

// bashRedirectOpWritesFile reports whether a shell redirect operator writes to
// a file. The Win7 build pins mvdan.cc/sh v3.7.0 (the newest release that
// compiles under go1.20.14); it predates the clobber operator constants
// (RdrClob/AppClob/RdrAllClob/AppAllClob, added in sh v3.13.0), so this variant
// matches only the write operators that v3.7.0 exposes. The affected forms
// (`>|`, `>>|`, `&>|`) are rare bash clobber overrides; missing them only
// slightly under-detects file-writing redirects for the reduced Win7 build.
func bashRedirectOpWritesFile(op syntax.RedirOperator) bool {
	switch op {
	case syntax.RdrOut, syntax.AppOut, syntax.RdrAll, syntax.AppAll, syntax.RdrInOut:
		return true
	default:
		return false
	}
}
