// Copyright © 2018 Shi Han NG
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/shihanng/bgist/gist"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var (
	public      bool
	description string
	accessToken string

	dummyFilename = "dummy.go"
	dummyContent  = "package dummy"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bgist",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MinimumNArgs(1),
	RunE: actual,
}

func actual(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	client := gist.NewClient(ctx, accessToken)

	info, err := client.CreateGist(ctx,
		gist.Description(description),
		gist.Public(public),
		gist.File(&github.GistFile{Filename: &dummyFilename, Content: &dummyContent}),
	)
	if err != nil {
		return err
	}

	fs := memfs.New()
	storer := memory.NewStorage()

	r, err := git.Clone(storer, fs, &git.CloneOptions{
		URL: info.GitURL,
	})
	if err != nil {
		return errors.Wrap(err, "clone gist to memory")
	}

	w, err := r.Worktree()
	if err != nil {
		return errors.Wrap(err, "getting the worktree")
	}

	for _, f := range args {
		filename, err := addFile(f, fs)
		if err != nil {
			return err
		}

		_, err = w.Add(filename)
		if err != nil {
			return errors.Wrap(err, "git add new filename")
		}
	}

	if _, err := w.Remove(string(dummyFilename)); err != nil {
		return errors.Wrap(err, "git rm")
	}

	_, err = w.Commit("", &git.CommitOptions{
		Author: &object.Signature{
			Name:  info.Name,
			Email: info.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return errors.Wrap(err, "git commit")
	}

	auth := &http.BasicAuth{Username: info.ID, Password: accessToken}

	if err := r.Push(&git.PushOptions{Auth: auth}); err != nil {
		return errors.Wrap(err, "git push")
	}

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&public, "public", false, "Publish as public gist")
	rootCmd.PersistentFlags().StringVarP(&description, "description", "d", "", "Description of the gist")

	viper.SetEnvPrefix("bgist")
	if err := viper.BindEnv("github_access_token"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	accessToken = viper.GetString("github_access_token")
	if accessToken == "" {
		fmt.Println(`GitHub's personal access token is needed as environment variable BGIST_GITHUB_ACCESS_TOKEN.`)
		fmt.Println(`It can be obtained from https://github.com/settings/tokens. The require scope is "gist".`)
		os.Exit(1)
	}
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
