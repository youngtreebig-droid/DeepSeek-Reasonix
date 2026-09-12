package compat

import (
	"crypto/rand"
	"reflect"
	"sync"
)

// This file backports a few standard-library additions that Reasonix uses but
// that only exist in Go 1.21+ / 1.22+ / 1.24+. The go1.20.14 toolchain used for
// the Windows 7 build cannot see them. Each helper is written without any
// Go 1.21+ language feature, so it compiles identically on both the native
// toolchain and go1.20.14, and callers use the helper on every build.

// TypeFor returns the reflect.Type of T. It mirrors reflect.TypeFor[T]
// (added in Go 1.22). Unlike reflect.TypeOf(x), it works for interface types
// and needs no value instance.
func TypeFor[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

// OnceValue returns a function that invokes f only once and returns the value
// returned by f on every call. It mirrors sync.OnceValue[T] (added in Go 1.21).
func OnceValue[T any](f func() T) func() T {
	var once sync.Once
	var value T
	return func() T {
		once.Do(func() { value = f() })
		return value
	}
}

// ClearSyncMap removes every entry from m. It mirrors (*sync.Map).Clear
// (added in Go 1.23) using Range+Delete so it compiles on go1.20.14. Deleting
// during Range is explicitly permitted by sync.Map.
func ClearSyncMap(m *sync.Map) {
	m.Range(func(key, _ any) bool {
		m.Delete(key)
		return true
	})
}

// WaitGroupGo runs f in a new goroutine tracked by wg, mirroring
// (*sync.WaitGroup).Go (added in Go 1.25). It compiles on go1.20.14.
func WaitGroupGo(wg *sync.WaitGroup, f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		f()
	}()
}

// randTextChars is the base32 alphabet crypto/rand.Text uses (RFC 4648,
// lowercase, no padding).
const randTextChars = "abcdefghijklmnopqrstuvwxyz234567"

// RandText returns a cryptographically random string of 26 base32 characters
// (about 130 bits of entropy), matching crypto/rand.Text (added in Go 1.24).
// It panics only if the system CSPRNG fails, mirroring the standard function.
func RandText() string {
	src := make([]byte, 26)
	if _, err := rand.Read(src); err != nil {
		panic(err)
	}
	for i := range src {
		src[i] = randTextChars[src[i]%32]
	}
	return string(src)
}
