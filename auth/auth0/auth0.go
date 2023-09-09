package auth0

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/interline-io/transitland-server/auth/authn"
)

type Auth0Client struct {
	client *management.Management
}

func NewAuth0Client(domain string, clientId string, clientSecret string) (*Auth0Client, error) {
	auth0API, err := management.New(
		domain,
		management.WithClientCredentials(clientId, clientSecret),
	)
	if err != nil {
		return nil, err
	}
	return &Auth0Client{client: auth0API}, nil
}

func (c *Auth0Client) UserByID(ctx context.Context, id string) (authn.User, error) {
	user, err := c.client.User.Read(id)
	if err != nil {
		return nil, err
	}
	u := authn.NewCtxUser(user.GetID(), user.GetName(), user.GetEmail())
	return u, nil
}

func (c *Auth0Client) Users(ctx context.Context, userQuery string) ([]authn.User, error) {
	ul, err := c.client.User.List(management.Query(userQuery))
	if err != nil {
		return nil, err
	}
	var ret []authn.User
	for _, user := range ul.Users {
		ret = append(ret, authn.NewCtxUser(user.GetID(), user.GetName(), user.GetEmail()))
	}
	return ret, nil
}