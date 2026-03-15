package security

import (
	"crypto/sha256"
	"encoding/hex"
)

type TokenHasher struct{}

func (h *TokenHasher) Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
