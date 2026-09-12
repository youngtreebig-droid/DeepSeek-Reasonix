//go:build !win7

package serve

import (
	"crypto/pbkdf2"
	"crypto/sha256"
)

// derivePBKDF2SessionKey derives a deterministic session key from the password
// hash using PBKDF2-HMAC-SHA256. It uses the std crypto/pbkdf2 package added in
// Go 1.24. The Win7 build (Go 1.20) uses golang.org/x/crypto/pbkdf2 instead;
// both produce identical keys for the same inputs.
func derivePBKDF2SessionKey(passwordHash string, salt []byte, iter, keyLen int) []byte {
	key, err := pbkdf2.Key(sha256.New, passwordHash, salt, iter, keyLen)
	if err != nil {
		panic("serve/auth: pbkdf2 failed: " + err.Error())
	}
	return key
}
