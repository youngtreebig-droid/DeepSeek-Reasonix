//go:build go1.21

package xslices

import "cmp"

// Ordered constrains the element types accepted by Sort and IsSorted.
type Ordered = cmp.Ordered
