package gist

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	billy "gopkg.in/src-d/go-billy.v4"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage"
)

func TestGit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockRepoer := NewMockrepoer(mockCtrl)

	var repo *git.Repository

	cloneFn = func(s storage.Storer, f billy.Filesystem, gitURL string) (
		repoer, *git.Worktree, error) {

		var err error
		repo, err = git.Init(s, f)
		if err != nil {
			return nil, nil, errors.Wrap(err, "when initing a repo")
		}

		w, err := repo.Worktree()
		if err != nil {
			return nil, nil, errors.Wrap(err, "when creating worktree")
		}

		return mockRepoer, w, nil
	}

	g, err := NewGit(testInfo, "secret")
	assert.NoError(t, err)

	assert.NoError(t, g.Add("./testdata/test_1.txt"))
	assert.NoError(t, g.Add("./testdata/test_2.txt"))
	assert.NoError(t, g.Commit("adding new files"))
	assert.NoError(t, g.Remove("test_1.txt"))
	assert.NoError(t, g.Commit("removing test_1.txt"))

	// Check if the changes are actually committed.
	cIter, err := repo.Log(&git.LogOptions{})
	require.NoError(t, err)
	defer cIter.Close()

	for _, expected := range []string{
		"removing test_1.txt",
		"adding new files",
	} {
		c, actualErr := cIter.Next()
		assert.NoError(t, actualErr)
		assert.Equal(t, expected, c.Message)
		assert.Equal(t, testInfo.Name, c.Author.Name)
		assert.Equal(t, testInfo.Email, c.Author.Email)
	}
	_, err = cIter.Next()
	assert.Error(t, err)

	mockRepoer.EXPECT().Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: testInfo.ID,
			Password: g.accessToken,
		},
	}).Return(nil)
	assert.NoError(t, g.Push())
}
