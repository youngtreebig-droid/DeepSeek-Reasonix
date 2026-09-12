//go:build go1.21

// Package xmaps re-exports the map helpers Reasonix uses. On the native
// toolchain the wrappers forward to the standard maps package; under the
// go1.20.14 Win7 toolchain they forward to golang.org/x/exp/maps.
//
// Note that maps.Keys differs between the two: the standard package returns a
// Go 1.23 iterator (iter.Seq[K]) while golang.org/x/exp/maps returns a slice.
// Keys is therefore only provided on the native side; the two call sites that
// use it are excluded from the Win7 build by a later feature.
package xmaps

import "maps"

func Clone[M ~map[K]V, K comparable, V any](m M) M { return maps.Clone(m) }

func Copy[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) { maps.Copy(dst, src) }

func Equal[M1, M2 ~map[K]V, K, V comparable](m1 M1, m2 M2) bool { return maps.Equal(m1, m2) }
