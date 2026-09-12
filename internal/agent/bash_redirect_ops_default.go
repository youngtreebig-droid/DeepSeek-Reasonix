//go:build !win7

package agent

import "mvdan.cc/sh/v3/syntax"

// bashRedirectOpWritesFile reports whether a shell redirect operator writes to
// a file (as opposed to reading or duplicating a descriptor). The full build
// uses the complete operator set from the current mvdan.cc/sh, including the
// clobber variants (>| >&| etc.) added in newer releases. The Win7 build pins
// an older sh that lacks those constants and matches the subset it has (see
// bash_redirect_ops_win7.go).
func bashRedirectOpWritesFile(op syntax.RedirOperator) bool {
	switch op {
	case syntax.RdrOut, syntax.AppOut, syntax.RdrClob, syntax.AppClob,
		syntax.RdrAll, syntax.RdrAllClob, syntax.AppAll, syntax.AppAllClob,
		syntax.RdrInOut:
		return true
	default:
		return false
	}
}
