package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTValidator struct {
	secret []byte
}

func NewJWTValidator(secret string) *JWTValidator {
	return &JWTValidator{secret: []byte(secret)}
}

func (j *JWTValidator) ValidateAccess(tokenStr string) (uuid.UUID, error) {
	var claims Claims

	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.ErrSignatureInvalid
		}
		return j.secret, nil
	})

	if err != nil {
		return uuid.Nil, errors.New("invalid token")
	}

	if !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	if claims.ID != "access" {
		return uuid.Nil, errors.New("invalid token type")
	}

	if claims.UserID == uuid.Nil {
		return uuid.Nil, errors.New("invalid token")
	}

	return claims.UserID, nil
}
