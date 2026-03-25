package security

type FakeHasher struct{}

func (f *FakeHasher) Hash(password string) (string, error) {
	return "hashed_Secure1111!!!!wwwwwwwsecure" + password, nil
}

// 🔥 ВАЖНО: (plain, hash)
func (f *FakeHasher) Compare(plain, hash string) bool {
	return "hashed_Secure1111!!!!wwwwwwwsecure"+plain == hash
}
