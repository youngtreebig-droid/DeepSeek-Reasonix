//go:build !go1.21

package xmaps

import "golang.org/x/exp/maps"

// Keys is intentionally omitted here: golang.org/x/exp/maps.Keys returns a
// slice whereas the standard maps.Keys returns a Go 1.23 iterator, and the call
// sites that use it rely on the iterator form. Those call sites are excluded
// from the Win7 build by a later feature.

func Clone[M ~map[K]V, K comparable, V any](m M) M { return maps.Clone(m) }

func Copy[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) { maps.Copy(dst, src) }

func Equal[M1, M2 ~map[K]V, K, V comparable](m1 M1, m2 M2) bool { return maps.Equal(m1, m2) }
