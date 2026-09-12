//go:build go1.23

package xmaps

import (
	"iter"
	"maps"
)

// Keys re-exports the Go 1.23 maps.Keys iterator helper. The
// golang.org/x/exp/maps.Keys equivalent returns a slice instead of an
// iterator, so Keys is only available on toolchains with standard iterator
// support. Call sites that use it are excluded from the Win7 build by a later
// feature.
func Keys[M ~map[K]V, K comparable, V any](m M) iter.Seq[K] { return maps.Keys(m) }
