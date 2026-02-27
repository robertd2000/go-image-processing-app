package security

type FakeHasher struct{}

func (f *FakeHasher) Hash(password string) (string, error) {
	return "hashed_" + password, nil
}

func (f *FakeHasher) Compare(hash, password string) bool {
	return hash == "hashed_"+password
}
