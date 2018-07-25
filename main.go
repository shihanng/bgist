package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"golang.org/x/oauth2"
)

var accessToken = os.Getenv("GITHUB_ACCESS_TOKEN")

var public bool
var description string

var dummyFilename github.GistFilename = "dummy.go"
var dummyContent = "package dummy"

func init() {
	if accessToken == "" {
		die("GITHUB_ACCESS_TOKEN is empty")
	}

	flag.BoolVar(&public, "public", false, "publish as public gist")
	flag.StringVarP(&description, "description", "d", "", "description of the gist")

	flag.ErrHelp = nil
}

func die(v interface{}) {
	fmt.Fprintf(os.Stderr, "%v\n", v)
	os.Exit(1)
}

func main() {
	flag.Parse()

	userID, err := createGist(description, public)
	if err != nil {
		die(err)
	}

	fmt.Println(userID)
}

func createGist(description string, public bool) (string, error) {
	ctx := context.Background()

	auth := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	))

	client := github.NewClient(auth)

	g := github.Gist{
		Description: &description,
		Public:      &public,
		Files: map[github.GistFilename]github.GistFile{
			dummyFilename: github.GistFile{
				Content: &dummyContent,
			},
		},
	}

	repos, _, err := client.Gists.Create(ctx, &g)
	if err != nil {
		return "", errors.Wrap(err, "create new gist")
	}

	return repos.GetOwner().GetLogin(), nil
}
