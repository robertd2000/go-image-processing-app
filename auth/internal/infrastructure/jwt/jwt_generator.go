// Package jwt
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	tokenDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
)

type CustomClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email,omitempty"`
	jwt.RegisteredClaims
}

type JWTGenerator struct {
	secret []byte
}

func NewJWTGenerator(secret []byte) *JWTGenerator {
	if len(secret) == 0 {
		panic("secret must not be empty")
	}
	return &JWTGenerator{secret: secret}
}

func (j *JWTGenerator) Generate(userID uuid.UUID, email string) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        "generic",
		},
	}

	return j.signedString(claims)
}

func (j *JWTGenerator) Validate(token string) (uuid.UUID, error) {
	claims, err := j.parseAndValidate(token, "generic")
	if err != nil {
		return uuid.Nil, err
	}
	if claims.Email == "" {
		return uuid.Nil, tokenDomain.ErrInvalidToken
	}
	return claims.UserID, nil
}

func (j *JWTGenerator) GenerateAccess(userID uuid.UUID) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        "access",
		},
	}
	return j.signedString(claims)
}

func (j *JWTGenerator) GenerateRefresh(userID uuid.UUID) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        "refresh",
		},
	}
	return j.signedString(claims)
}

func (j *JWTGenerator) ValidateAccess(token string) (uuid.UUID, error) {
	claims, err := j.parseAndValidate(token, "access")
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

func (j *JWTGenerator) ValidateRefresh(token string) (uuid.UUID, error) {
	claims, err := j.parseAndValidate(token, "refresh")
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

func (j *JWTGenerator) signedString(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTGenerator) parseAndValidate(tokenStr, expectedType string) (*CustomClaims, error) {
	var claims CustomClaims

	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.ErrSignatureInvalid
		}
		return j.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, tokenDomain.ErrInvalidToken
		}
		return nil, tokenDomain.ErrInvalidToken
	}

	if !token.Valid {
		return nil, tokenDomain.ErrInvalidToken
	}

	if claims.ID != expectedType {
		return nil, tokenDomain.ErrInvalidToken
	}

	if claims.UserID == uuid.Nil {
		return nil, tokenDomain.ErrInvalidToken
	}

	return &claims, nil
}
