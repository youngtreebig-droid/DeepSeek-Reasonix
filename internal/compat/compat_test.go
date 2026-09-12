package compat

import (
	"math"
	"testing"
)

func TestMinMax(t *testing.T) {
	if got := Min(3, 5); got != 3 {
		t.Errorf("Min(3,5) = %d, want 3", got)
	}
	if got := Max(3, 5); got != 5 {
		t.Errorf("Max(3,5) = %d, want 5", got)
	}
	if got := Min("b", "a"); got != "a" {
		t.Errorf("Min(b,a) = %q, want a", got)
	}
	if got := Max(2.5, 2.5); got != 2.5 {
		t.Errorf("Max(2.5,2.5) = %v, want 2.5", got)
	}
}

func TestClear(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	Clear(m)
	if len(m) != 0 {
		t.Errorf("Clear left %d entries, want 0", len(m))
	}
}

func TestClearSlice(t *testing.T) {
	s := []int{1, 2, 3}
	ClearSlice(s)
	for i, v := range s {
		if v != 0 {
			t.Errorf("ClearSlice s[%d] = %d, want 0", i, v)
		}
	}
	if len(s) != 3 {
		t.Errorf("ClearSlice changed length to %d, want 3", len(s))
	}
}

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b, want int
	}{
		{1, 2, -1},
		{2, 2, 0},
		{3, 2, +1},
	}
	for _, c := range cases {
		if got := Compare(c.a, c.b); got != c.want {
			t.Errorf("Compare(%d,%d) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
	nan := math.NaN()
	if got := Compare(nan, nan); got != 0 {
		t.Errorf("Compare(NaN,NaN) = %d, want 0", got)
	}
	if got := Compare(nan, 1.0); got != -1 {
		t.Errorf("Compare(NaN,1) = %d, want -1", got)
	}
	if got := Compare(1.0, nan); got != +1 {
		t.Errorf("Compare(1,NaN) = %d, want +1", got)
	}
}
