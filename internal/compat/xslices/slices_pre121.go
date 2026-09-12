//go:build !go1.21

package xslices

import "golang.org/x/exp/slices"

// The Go 1.23 iterator helpers slices.Backward and slices.Sorted have no
// golang.org/x/exp/slices equivalent and are intentionally omitted here. Call
// sites that use them are excluded from the Win7 build by a later feature.

func Clone[S ~[]E, E any](s S) S { return slices.Clone(s) }

func Compact[S ~[]E, E comparable](s S) S { return slices.Compact(s) }

func Contains[S ~[]E, E comparable](s S, v E) bool { return slices.Contains(s, v) }

func ContainsFunc[S ~[]E, E any](s S, f func(E) bool) bool { return slices.ContainsFunc(s, f) }

func DeleteFunc[S ~[]E, E any](s S, del func(E) bool) S { return slices.DeleteFunc(s, del) }

func Equal[S ~[]E, E comparable](s1, s2 S) bool { return slices.Equal(s1, s2) }

func Index[S ~[]E, E comparable](s S, v E) int { return slices.Index(s, v) }

func IndexFunc[S ~[]E, E any](s S, f func(E) bool) int { return slices.IndexFunc(s, f) }

func IsSorted[S ~[]E, E Ordered](s S) bool { return slices.IsSorted(s) }

func Reverse[S ~[]E, E any](s S) { slices.Reverse(s) }

func Sort[S ~[]E, E Ordered](s S) { slices.Sort(s) }

func SortFunc[S ~[]E, E any](s S, cmp func(a, b E) int) { slices.SortFunc(s, cmp) }

func SortStableFunc[S ~[]E, E any](s S, cmp func(a, b E) int) { slices.SortStableFunc(s, cmp) }
