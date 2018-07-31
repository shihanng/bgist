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
	"errors"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"github.com/shihanng/bgist/gist"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	Use:     "bgist",
	Example: "BGIST_GITHUB_ACCESS_TOKEN=secret bgist -d \"a demo\" photo-1.png photo-2.jpg",
	Short:   "A tool to upload image/binary file to gist.github.com.",
	Long: `A tool to upload image/binary file to gist.github.com.

GitHub's personal access token should be provided as BGIST_GITHUB_ACCESS_TOKEN
for this tool to work.

https://github.com/shihanng/bgist`,
	Args: cobra.MinimumNArgs(1),
	RunE: actual,
}

func actual(cmd *cobra.Command, args []string) error {
	if accessToken == "" {
		fmt.Println(`GitHub's personal access token is needed as environment variable BGIST_GITHUB_ACCESS_TOKEN.`)
		fmt.Println(`It can be obtained from https://github.com/settings/tokens. The require scope is "gist".`)
		return errors.New("BGIST_GITHUB_ACCESS_TOKEN is empty")
	}

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
	fmt.Println("Created", info.HTMLURL)

	g, err := gist.NewGit(info, accessToken)
	if err != nil {
		return err
	}

	for _, f := range args {
		if err := g.Add(f); err != nil {
			return err
		}
	}

	if err := g.Remove(dummyFilename); err != nil {
		return err
	}

	if err := g.Commit("update"); err != nil {
		return err
	}

	if err := g.Push(); err != nil {
		return err
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
}
