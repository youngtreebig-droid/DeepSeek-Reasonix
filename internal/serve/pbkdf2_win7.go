//go:build win7

package serve

import (
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
)

// derivePBKDF2SessionKey derives a deterministic session key from the password
// hash using PBKDF2-HMAC-SHA256. The std crypto/pbkdf2 package only exists in
// Go 1.24+, which the go1.20.14 Win7 toolchain cannot build, so this variant
// uses golang.org/x/crypto/pbkdf2. The x/crypto helper never returns an error;
// it produces the same key as the std package for identical inputs.
func derivePBKDF2SessionKey(passwordHash string, salt []byte, iter, keyLen int) []byte {
	return pbkdf2.Key([]byte(passwordHash), salt, iter, keyLen, sha256.New)
}
