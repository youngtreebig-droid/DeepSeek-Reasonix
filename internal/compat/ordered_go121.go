//go:build go1.21

package compat

import "cmp"

// Ordered is the set of types that support the < <= >= > operators. On the
// native toolchain (Go 1.21+) it aliases the standard cmp.Ordered constraint.
type Ordered = cmp.Ordered
