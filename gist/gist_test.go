package gist

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	github "github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

func newString(v string) *string {
	return &v
}

func newBool(v bool) *bool {
	return &v
}

func TestClient(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGister := NewMockGister(mockCtrl)

	ctx := context.Background()
	c := NewClient(ctx, "")
	c.gist = mockGister

	testDescription := "A description"
	testGistFile1 := github.GistFile{
		Filename: &tmpFilename,
		Content:  &tmpContent,
	}
	testGistFile2 := github.GistFile{
		Filename: newString("test.txt"),
		Content:  newString("test"),
	}

	expected := Info{
		ID:      "johndoe",
		Name:    "John Doe",
		Email:   "jdoe@example.com",
		HTMLURL: "https://gist.github.com/johndoe/abc123",
		GitURL:  "git@gist.github.com:abc123.git",
	}

	mockGister.EXPECT().Create(ctx, &github.Gist{
		Description: &testDescription,
		Public:      newBool(true),
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(*testGistFile1.Filename): testGistFile1,
			github.GistFilename(*testGistFile2.Filename): testGistFile2,
		},
	}).Return(&github.Gist{
		Owner: &github.User{
			Login: &expected.ID,
			Name:  &expected.Name,
			Email: &expected.Email,
		},
		HTMLURL:    &expected.HTMLURL,
		GitPullURL: &expected.GitURL,
	}, nil, nil)

	actual, err := c.CreateGist(ctx,
		Description(testDescription),
		Public(true),
		File(&testGistFile1),
		File(&testGistFile2),
	)

	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(expected, actual)
}
