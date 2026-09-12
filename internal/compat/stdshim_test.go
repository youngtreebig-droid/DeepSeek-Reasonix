package compat

import (
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestRandText(t *testing.T) {
	const iterations = 2000

	// Membership: every rune must come from the base32 alphabet, and the
	// length must be exactly 26 (the crypto/rand.Text contract, ~130 bits).
	allowed := map[rune]bool{}
	for _, c := range randTextChars {
		allowed[c] = true
	}
	if len(randTextChars) != 32 {
		t.Fatalf("randTextChars has %d chars, want 32 (base32)", len(randTextChars))
	}

	seen := make(map[string]struct{}, iterations)
	charCounts := map[rune]int{}
	for i := 0; i < iterations; i++ {
		s := RandText()
		if got := len([]rune(s)); got != 26 {
			t.Fatalf("RandText() length = %d, want 26 (value %q)", got, s)
		}
		for _, r := range s {
			if !allowed[r] {
				t.Fatalf("RandText() produced rune %q not in alphabet %q", r, randTextChars)
			}
			charCounts[r]++
		}
		if _, dup := seen[s]; dup {
			t.Fatalf("RandText() produced a duplicate value %q within %d draws", s, iterations)
		}
		seen[s] = struct{}{}
	}

	// Basic distribution sanity: across 2000*26 = 52000 characters over a
	// 32-symbol alphabet the expected count per symbol is ~1625. Every symbol
	// should appear at least once; a broken mapping (e.g. modulo bias dropping
	// symbols, or a constant output) would fail this.
	if len(charCounts) != 32 {
		t.Errorf("RandText() used %d distinct symbols, want all 32", len(charCounts))
	}
	total := iterations * 26
	expected := float64(total) / 32.0
	for _, c := range randTextChars {
		got := charCounts[c]
		// Allow a generous band (±60%) to avoid flakiness while still catching
		// a symbol that is systematically over/under-represented.
		if float64(got) < expected*0.4 || float64(got) > expected*1.6 {
			t.Errorf("symbol %q appeared %d times, expected near %.0f (skewed distribution)", c, got, expected)
		}
	}
}

func TestOnceValue(t *testing.T) {
	var calls int32
	fn := OnceValue(func() int {
		atomic.AddInt32(&calls, 1)
		return 42
	})

	// Value is stable across calls and f runs exactly once even under
	// concurrent access.
	var wg sync.WaitGroup
	results := make([]int, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = fn()
		}(i)
	}
	wg.Wait()

	for i, v := range results {
		if v != 42 {
			t.Errorf("OnceValue result[%d] = %d, want 42", i, v)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("OnceValue invoked f %d times, want exactly 1", got)
	}
}

func TestTypeFor(t *testing.T) {
	if got := TypeFor[int]().String(); got != "int" {
		t.Errorf("TypeFor[int]() = %q, want int", got)
	}
	if got := TypeFor[[]string]().String(); got != "[]string" {
		t.Errorf("TypeFor[[]string]() = %q, want []string", got)
	}
	// TypeFor must work for interface types, where reflect.TypeOf(x) would need
	// a concrete instance.
	type stringerish interface{ String() string }
	got := TypeFor[stringerish]()
	if got.Kind() != reflect.Interface {
		t.Errorf("TypeFor[interface] Kind = %v, want interface", got.Kind())
	}
}

func TestClearSyncMap(t *testing.T) {
	var m sync.Map
	for i := 0; i < 10; i++ {
		m.Store(i, i*i)
	}
	ClearSyncMap(&m)

	count := 0
	m.Range(func(_, _ any) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("ClearSyncMap left %d entries, want 0", count)
	}

	// Map is still usable after clearing.
	m.Store("k", "v")
	if v, ok := m.Load("k"); !ok || v != "v" {
		t.Errorf("sync.Map unusable after ClearSyncMap: got %v, ok=%v", v, ok)
	}
}

func TestWaitGroupGo(t *testing.T) {
	var wg sync.WaitGroup
	var counter int32
	for i := 0; i < 20; i++ {
		WaitGroupGo(&wg, func() {
			atomic.AddInt32(&counter, 1)
		})
	}
	wg.Wait()
	if got := atomic.LoadInt32(&counter); got != 20 {
		t.Errorf("WaitGroupGo ran %d goroutines, want 20", got)
	}
}

// TestRandTextNoModuloBias documents the invariant the review flagged: the
// alphabet length must evenly divide 256 so the `% 32` mapping is unbiased. If
// a future edit widens the alphabet to a size that does not divide 256, this
// fails, forcing the author to use a rejection-sampling scheme instead.
func TestRandTextNoModuloBias(t *testing.T) {
	if 256%len(randTextChars) != 0 {
		t.Fatalf("len(randTextChars)=%d does not divide 256; %% mapping is biased", len(randTextChars))
	}
	if strings.ContainsAny(randTextChars, "018") {
		// base32 (RFC 4648) intentionally omits 0/1/8; this guards the alphabet
		// identity against accidental edits.
		t.Errorf("randTextChars %q contains a non-base32 symbol", randTextChars)
	}
}
