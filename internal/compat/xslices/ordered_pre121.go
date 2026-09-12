//go:build !go1.21

package xslices

import "golang.org/x/exp/constraints"

// Ordered constrains the element types accepted by Sort and IsSorted.
type Ordered = constraints.Ordered
