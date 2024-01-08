package google

import (
	"context"
	"encoding/json"
	"github.com/cicconee/clox/internal/oauth2"
	"github.com/cicconee/clox/internal/provider"
)

const UserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"

type Client struct {
	http provider.HTTPClient
}

func New(client provider.HTTPClient) *Client {
	return &Client{http: client}
}

type User struct {
	Sub        string `json:"sub"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	PictureURL string `json:"picture"`
	Email      string `json:"email"`
}

func (c *Client) UserInfo(ctx context.Context, token *oauth2.Token) (provider.User, error) {
	client := c.http.Client(ctx, token)
	resp, err := client.Get(UserInfoURL)
	if err != nil {
		return provider.User{}, err
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return provider.User{}, err
	}

	return provider.User{
		ID:         provider.NewID(token.Provider(), user.Sub),
		FirstName:  user.GivenName,
		LastName:   user.FamilyName,
		PictureURL: user.PictureURL,
		Email:      user.Email,
	}, nil
}
