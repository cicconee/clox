package provider

import (
	"context"
	"fmt"
	"github.com/cicconee/clox/internal/oauth2"
	"net/http"
	"strings"
)

type HTTPClient interface {
	Client(context.Context, *oauth2.Token) *http.Client
}

type User struct {
	ID         ID     `json:"sub"`
	FirstName  string `json:"given_name"`
	LastName   string `json:"family_name"`
	PictureURL string `json:"picture"`
	Email      string `json:"email"`
}

type ID struct {
	Provider string
	Sub      string
}

func NewID(provider string, sub string) ID {
	return ID{Provider: provider, Sub: sub}
}

func Decode(s string) (ID, error) {
	strs := strings.Split(s, "|")
	if len(strs) != 2 {
		return ID{}, fmt.Errorf("invalid format: not formatted as provider|sub")
	}

	return ID{Provider: strs[0], Sub: strs[1]}, nil
}

func (i *ID) Encode() string {
	return fmt.Sprintf("%s|%s", i.Provider, i.Sub)
}
