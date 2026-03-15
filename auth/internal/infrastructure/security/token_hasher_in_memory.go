package security

type FakeTokenHasher struct{}

func (f *FakeTokenHasher) Hash(token string) string {
	return "hashed_" + token
}
