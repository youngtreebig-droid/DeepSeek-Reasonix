//go:build go1.23

package xslices

import (
	"iter"
	"slices"
)

// Backward re-exports the Go 1.23 slices.Backward iterator helper. It has no
// golang.org/x/exp/slices equivalent, so it is only available on toolchains
// with the standard iterator support. Call sites that use it are excluded from
// the Win7 build by a later feature.
func Backward[S ~[]E, E any](s S) iter.Seq2[int, E] { return slices.Backward(s) }

// Sorted re-exports the Go 1.23 slices.Sorted iterator helper.
func Sorted[E Ordered](seq iter.Seq[E]) []E { return slices.Sorted(seq) }
