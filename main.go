package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"golang.org/x/oauth2"
	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var (
	accessToken = os.Getenv("BGIST_GITHUB_ACCESS_TOKEN")

	public      bool
	description string

	dummyFilename github.GistFilename = "dummy.go"
	dummyContent                      = "package dummy"
)

func init() {
	if accessToken == "" {
		die("BGIST_GITHUB_ACCESS_TOKEN is empty")
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

	if flag.NArg() < 1 {
		die("need at least one file as argument")
	}

	user, gitPullURL, err := createGist(description, public)
	if err != nil {
		die(err)
	}

	fs := memfs.New()
	storer := memory.NewStorage()

	r, err := git.Clone(storer, fs, &git.CloneOptions{
		URL: gitPullURL,
	})
	if err != nil {
		die(errors.Wrap(err, "clone gist to memory"))
	}

	w, err := r.Worktree()
	if err != nil {
		die(errors.Wrap(err, "getting the worktree"))
	}

	for _, f := range flag.Args() {
		filename, err := addFile(f, fs)
		if err != nil {
			die(err)
		}

		_, err = w.Add(filename)
		if err != nil {
			die(errors.Wrap(err, "git add new filename"))
		}
	}

	if _, err := w.Remove(string(dummyFilename)); err != nil {
		die(errors.Wrap(err, "git rm"))
	}

	_, err = w.Commit("", &git.CommitOptions{
		Author: &object.Signature{
			Name:  user.GetName(),
			Email: user.GetEmail(),
			When:  time.Now(),
		},
	})
	if err != nil {
		die(errors.Wrap(err, "git commit"))
	}

	auth := &http.BasicAuth{Username: user.GetLogin(), Password: accessToken}

	if err := r.Push(&git.PushOptions{Auth: auth}); err != nil {
		die(errors.Wrap(err, "git push"))
	}
}

func createGist(description string, public bool) (*github.User, string, error) {
	ctx := context.Background()

	auth := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
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

	gists, _, err := client.Gists.Create(ctx, &g)
	if err != nil {
		return nil, "", errors.Wrap(err, "create new gist")
	}
	fmt.Println("created:", gists.GetHTMLURL())

	return gists.GetOwner(), gists.GetGitPullURL(), nil
}

func addFile(source string, destination billy.Filesystem) (string, error) {
	filename := filepath.Base(source)

	newFile, err := destination.Create(filename)
	if err != nil {
		return "", errors.Wrap(err, "create new file in gist fs")
	}
	defer newFile.Close()

	sourceFile, err := os.Open(source)
	if err != nil {
		return "", errors.Wrap(err, "open new file")
	}
	defer sourceFile.Close()

	_, err = io.Copy(newFile, sourceFile)
	if err != nil {
		return "", errors.Wrap(err, "copy the file to gist fs")
	}

	return filename, nil
}
