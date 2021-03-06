//go:generate mockgen -source=git.go -destination=mock_git_test.go -package=gist
package gist

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var cloneFn = func(s storage.Storer, f billy.Filesystem, gitURL string) (
	repoer, *git.Worktree, error) {

	r, err := git.Clone(s, f, &git.CloneOptions{
		URL: gitURL,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "when cloning a repo")
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, nil, errors.Wrap(err, "when creating worktree")
	}

	return r, w, nil
}

type repoer interface {
	Worktree() (*git.Worktree, error)
	Log(*git.LogOptions) (object.CommitIter, error)
	Push(*git.PushOptions) error
}

type Git struct {
	info        Info
	accessToken string

	filesystem billy.Filesystem
	storage    *memory.Storage

	repo     repoer
	worktree *git.Worktree
}

func NewGit(info Info, accessToken string) (*Git, error) {
	f := memfs.New()
	s := memory.NewStorage()

	r, w, err := cloneFn(s, f, info.GitURL)
	if err != nil {
		return nil, err
	}

	return &Git{
		info:        info,
		accessToken: accessToken,

		filesystem: f,
		storage:    s,

		repo:     r,
		worktree: w,
	}, nil
}

func (g *Git) Add(path string) error {
	filename := filepath.Base(path)

	newFile, err := g.filesystem.Create(filename)
	if err != nil {
		return errors.Wrap(err, "when creating a new file in filesystem")
	}
	defer newFile.Close()

	sourceFile, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "when openning the source")
	}
	defer sourceFile.Close()

	if _, err = io.Copy(newFile, sourceFile); err != nil {
		return errors.Wrap(err, "when copying the source to filesystem")
	}

	_, err = g.worktree.Add(filename)
	return errors.Wrap(err, "when adding new file to repo")
}

func (g *Git) Remove(filename string) error {
	_, err := g.worktree.Remove(filename)
	return errors.Wrap(err, "when removing file from repo")
}

func (g *Git) Commit(msg string) error {
	_, err := g.worktree.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  g.info.Name,
			Email: g.info.Email,
			When:  time.Now(),
		},
	})
	return errors.Wrap(err, "when commiting")
}

func (g *Git) Push() error {
	auth := &http.BasicAuth{Username: g.info.ID, Password: g.accessToken}
	return errors.Wrap(g.repo.Push(&git.PushOptions{Auth: auth}), "when pushing")
}
