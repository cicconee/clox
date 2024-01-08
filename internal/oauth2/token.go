package oauth2

import (
	"golang.org/x/oauth2"
	"time"
)

type Token struct {
	oauth2Token *oauth2.Token
	provider    string
}

func (t *Token) AccessToken() string {
	return t.oauth2Token.AccessToken
}

func (t *Token) TokenType() string {
	return t.oauth2Token.TokenType
}

func (t *Token) RefreshToken() string {
	return t.oauth2Token.RefreshToken
}

func (t *Token) Expiry() time.Time {
	return t.oauth2Token.Expiry
}

func (t *Token) Provider() string {
	return t.provider
}
