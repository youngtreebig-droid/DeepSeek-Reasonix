// Package compat provides small generic helpers that replace the Go 1.21
// builtins (min, max, clear) and the Go 1.21 cmp.Compare function so the same
// source compiles on both the native toolchain and the go1.20.14 toolchain used
// for the Windows 7 build.
//
// The builtins min/max/clear and the cmp standard package were introduced in
// Go 1.21. Under Go 1.20 they do not exist, so the Win7 downgrade routes every
// use through these helpers. The helpers are defined without relying on any
// Go 1.21 language feature, so they compile identically on both toolchains.
package compat

// Min returns the smaller of a and b. It mirrors the Go 1.21 builtin min for
// two ordered operands.
func Min[T Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of a and b. It mirrors the Go 1.21 builtin max for
// two ordered operands.
func Max[T Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Clear removes every entry from the map m, mirroring the Go 1.21 builtin
// clear for maps.
func Clear[M ~map[K]V, K comparable, V any](m M) {
	for k := range m {
		delete(m, k)
	}
}

// ClearSlice zeroes every element of s, mirroring the Go 1.21 builtin clear for
// slices. It is a distinct name from Clear because Go generics cannot dispatch
// a single identifier across both map and slice arguments.
func ClearSlice[S ~[]E, E any](s S) {
	var zero E
	for i := range s {
		s[i] = zero
	}
}

// Compare returns -1, 0 or +1 depending on whether a is less than, equal to or
// greater than b. It mirrors the Go 1.21 cmp.Compare function. NaN is treated
// as less than any non-NaN value and equal to any other NaN, matching the
// standard library semantics.
func Compare[T Ordered](a, b T) int {
	xNaN := isNaN(a)
	yNaN := isNaN(b)
	if xNaN && yNaN {
		return 0
	}
	if xNaN || a < b {
		return -1
	}
	if yNaN || a > b {
		return +1
	}
	return 0
}

// isNaN reports whether x is not-a-number without importing the math package,
// which keeps this helper usable for any ordered type.
func isNaN[T Ordered](x T) bool {
	return x != x
}
