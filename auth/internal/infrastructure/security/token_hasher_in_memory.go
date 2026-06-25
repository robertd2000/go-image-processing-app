package security

func (h *TokenHasher) HashInMemory(token string) string {
	return h.Hash(token)
}
