package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/robertd2000/go-image-processing-app/auth/internal/domain/auth"
	tokenDomain "github.com/robertd2000/go-image-processing-app/auth/internal/domain/token"
	"github.com/robertd2000/go-image-processing-app/auth/internal/usecase/auth/model"
)

type JWTGenerator struct {
	secret []byte
}

func NewJWTGenerator(secret []byte) *JWTGenerator {
	if len(secret) == 0 {
		panic("secret must not be empty")
	}
	return &JWTGenerator{secret: secret}
}

// ACCESS TOKEN

func (j *JWTGenerator) GenerateAccess(input model.ClaimsInput) (string, error) {
	claims := CustomClaims{
		UserID:    input.UserID,
		Roles:     input.Roles,
		TokenType: TokenAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-service",
			Audience:  []string{"image-service", "user-service"},
		},
	}

	return j.signedString(claims)
}

func (j *JWTGenerator) ValidateAccess(token string) (*auth.Claims, error) {
	claims, err := j.parseAndValidate(token, TokenAccess)
	if err != nil {
		return nil, err
	}

	return toDomainClaims(*claims), nil
}

// REFRESH TOKEN

func (j *JWTGenerator) GenerateRefresh(userID uuid.UUID) (string, error) {
	claims := CustomClaims{
		UserID:    userID,
		TokenType: TokenRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-service",
		},
	}

	return j.signedString(claims)
}

func (j *JWTGenerator) ValidateRefresh(token string) (uuid.UUID, error) {
	claims, err := j.parseAndValidate(token, TokenRefresh)
	if err != nil {
		return uuid.Nil, err
	}

	return claims.UserID, nil
}

//  helpers

func (j *JWTGenerator) signedString(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTGenerator) parseAndValidate(tokenStr string, expectedType TokenType) (*CustomClaims, error) {
	var claims CustomClaims

	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return j.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, tokenDomain.ErrExpiredToken
		}
		return nil, tokenDomain.ErrInvalidToken
	}

	if !token.Valid {
		return nil, tokenDomain.ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return nil, tokenDomain.ErrInvalidToken
	}

	if claims.UserID == uuid.Nil {
		return nil, tokenDomain.ErrInvalidToken
	}

	if claims.Issuer != "auth-service" {
		return nil, tokenDomain.ErrInvalidToken
	}

	return &claims, nil
}
