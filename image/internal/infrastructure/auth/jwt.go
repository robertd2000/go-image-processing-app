package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/robertd2000/go-image-processing-app/image/internal/port"
)

type jwtClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Roles  []string  `json:"roles"`
	jwt.RegisteredClaims
}

type JWTValidator struct {
	secret []byte
}

func NewJWTValidator(secret string) *JWTValidator {
	return &JWTValidator{secret: []byte(secret)}
}

func (j *JWTValidator) ValidateAccess(tokenStr string) (*port.AuthClaims, error) {
	var claims jwtClaims

	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.ErrSignatureInvalid
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, errors.New("invalid token")
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.UserID == uuid.Nil {
		return nil, errors.New("invalid token")
	}

	return &port.AuthClaims{
		UserID: claims.UserID,
		Roles:  claims.Roles,
	}, nil
}
