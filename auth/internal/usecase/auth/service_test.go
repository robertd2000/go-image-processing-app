package auth_test

type (
	AuthService interface{}

	AuthTestSuite struct {
		service AuthService
	}
)

//
// func (s *AuthTestSuite) SetupTest() {
// 	s.service = auth.NewAuthService(userRepo user.UserRepository, refreshRepo token.TokenRepository, hasher port.PasswordHasher, tokenGen port.TokenGenerator)
// }
