//go:build !go1.21

package compat

import "golang.org/x/exp/constraints"

// Ordered is the set of types that support the < <= >= > operators. Under the
// go1.20.14 Win7 toolchain the standard cmp package does not exist, so it uses
// the equivalent constraint from golang.org/x/exp/constraints.
type Ordered = constraints.Ordered
