package token

import "time"

// NewListing is a token and its listing information. NewListing is returned when a new token
// is created.
type NewListing struct {
	// The signed JWT as a string.
	Token string

	// The listing information. This data is useful for the end user.
	Listing
}

// Listing is the relevant token information for the end user.
type Listing struct {
	// The token ID (jti).
	TokenID string

	// The token name. This field was defined by the user when creating the token.
	TokenName string

	// The expiration time (exp) of the token in UTC time.
	ExpiresAt time.Time

	// The issued at time (iat) of the token in UTC time.
	IssuedAt time.Time

	// The last time this token was used in UTC time.
	LastUsed time.Time
}

// Returns this Listing's ExpiresAt field as a string formatted as "2006-01-02T15:04:05Z07:00".
func (l *Listing) ExpiresAtString() string {
	return l.ExpiresAt.Format(time.RFC3339)
}

// Returns this Listing's IssuedAt field as a string formatted as "2006-01-02T15:04:05Z07:00".
func (l *Listing) IssuedAtString() string {
	return l.IssuedAt.Format(time.RFC3339)
}

// Returns this Listings LastUsed field as a string formatted as "2006-01-02T15:04:05Z07:00".
func (l *Listing) LastUsedString() string {
	return l.LastUsed.Format(time.RFC3339)
}
