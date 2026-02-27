package token

import "errors"

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrInvalidRefresh = errors.New("invalid refresh")
	ErrExpiredToken   = errors.New("expired token")
)

type Tokens struct {
	accessToken  string
	refreshToken string
}

func NewTokens(accessToken, refreshToken string) (*Tokens, error) {
	if err := validarteToken(accessToken); err != nil {
		return nil, err
	}

	if err := validarteToken(refreshToken); err != nil {
		return nil, err
	}

	return &Tokens{
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}, nil
}

func (t *Tokens) GetAccessToken() string {
	return t.accessToken
}

func (t *Tokens) GetRefreshToken() string {
	return t.refreshToken
}

func validarteToken(token string) error {
	if token == "" {
		return ErrInvalidToken
	}
	return nil
}
