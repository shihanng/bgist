//go:generate mockgen -source=gist.go -destination=mock_test.go -package=gist
package gist

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

var (
	tmpFilename = "dummy.go"
	tmpContent  = "package dummy"
)

// Gister is used to create gist on GitHub.
type Gister interface {
	Create(context.Context, *github.Gist) (*github.Gist, *github.Response, error)
}

// Client should be created with NewClient.
type Client struct {
	gist Gister
}

// NewClient created the client to create gist on GitHub.
func NewClient(ctx context.Context, accessToken string) *Client {
	auth := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	))

	client := github.NewClient(auth)

	return &Client{
		gist: client.Gists,
	}
}

// Info of the newly created gist.
type Info struct {
	ID      string
	Name    string
	Email   string
	HTMLURL string
	GitURL  string
}

// CreateGist creates the gist on GitHub based on the provided option.
func (c *Client) CreateGist(ctx context.Context, ops ...Option) (Info, error) {
	var g github.Gist

	for _, o := range ops {
		o(&g)
	}

	created, _, err := c.gist.Create(ctx, &g)
	if err != nil {
		return Info{}, errors.Wrap(err, "when creating new gist")
	}

	return Info{
		ID:      created.GetOwner().GetLogin(),
		Name:    created.GetOwner().GetName(),
		Email:   created.GetOwner().GetEmail(),
		HTMLURL: created.GetHTMLURL(),
		GitURL:  created.GetGitPullURL(),
	}, nil
}

// Option for CreateGist.
type Option func(*github.Gist)

// Description of the gist.
func Description(v string) Option {
	return func(g *github.Gist) {
		g.Description = &v
	}
}

// Public if set true, the gist will be created as public gist.
func Public(v bool) Option {
	return func(g *github.Gist) {
		g.Public = &v
	}
}

// File that will be uploaded when a new gist is created.
func File(gf *github.GistFile) Option {
	return func(g *github.Gist) {
		if g.Files == nil {
			g.Files = map[github.GistFilename]github.GistFile{
				github.GistFilename(*gf.Filename): *gf,
			}
			return
		}

		g.Files[github.GistFilename(*gf.Filename)] = *gf
	}
}

func (c *Client) ModifyGistRepo(url string) error {
	return nil
}
