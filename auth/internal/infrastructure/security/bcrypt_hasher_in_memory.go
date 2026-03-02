package security

type FakeHasher struct{}

func (f *FakeHasher) Hash(password string) (string, error) {
	return "hashed_Secure1111!!!!wwwwwwwsecure" + password, nil
}

func (f *FakeHasher) Compare(hash, password string) bool {
	return hash == "hashed_Secure1111!!!!wwwwwwwsecure"+password
}
