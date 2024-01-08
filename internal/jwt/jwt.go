package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Manager creates and validates JWT tokens. If using multiple sets of JWT's make sure to use multiple
// instances of Manager. The Manager fields will be used when both creating and validating, so it is
// important to validate the tokens with the same Manager values as when creating.
type Manager struct {
	// The secret key the manager will use to sign JWT's.
	secret string

	// The principle that issued the JWT. This will be the value in the tokens 'iss' claim.
	issuer string

	// The recipient that the JWT is intended for. This is the value in the tokens 'aud' claim.
	audience string
}

// NewManager creates a new JWT Manager.
func NewManager(issuer string, audience string) *Manager {
	return &Manager{issuer: issuer, audience: audience}
}

// SetSecret sets the secret key for this manager.
func (m *Manager) SetSecret(secret string) {
	m.secret = secret
}

// Claims is the JWT claims.
type Claims struct {
	jwt.RegisteredClaims
}

// NewTokenClaims holds the claims used when creating a new JWT.
type NewTokenClaims struct {
	Sub string
	Exp time.Time
	Nbf time.Time
	Iat time.Time
	Jti string
}

// New creates a JWT and returns it as a string.
//
// The token claims sub, exp, nbf, iat, and jti are set to the NewTokenClaims fields. The iss and aud claims
// are set to this managers audience and issuer fields.
//
// Tokens are signed with this managers secret.
func (m *Manager) New(c NewTokenClaims) (string, error) {
	claims := &Claims{RegisteredClaims: jwt.RegisteredClaims{
		Issuer:    m.issuer,
		Subject:   c.Sub,
		Audience:  jwt.ClaimStrings{m.audience},
		ExpiresAt: jwt.NewNumericDate(c.Exp),
		NotBefore: jwt.NewNumericDate(c.Nbf),
		IssuedAt:  jwt.NewNumericDate(c.Iat),
		ID:        c.Jti,
	}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(m.secret))
}

// Validate validates a JWT and returns the token claims.
//
// Tokens are validated with this managers secret. Validate also checks that the issuer (iss) and audience
// (aud) match this managers issuer and audience fields.
func (m *Manager) Validate(token string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{},
		m.keyFunc,
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.audience))
	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(*Claims)
	if !ok && t.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// keyFunc is a jwt.KeyFunc that is used when parsing a token.
func (m *Manager) keyFunc(t *jwt.Token) (interface{}, error) {
	return []byte(m.secret), nil
}
