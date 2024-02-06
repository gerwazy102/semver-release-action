package git

import (
	"context"
	"net/http"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/gerwazy102/semver-release-action/internal/pkg/action"
	"github.com/gerwazy102/semver-release-action/internal/pkg/enterprise"
	"github.com/google/go-github/v45/github"
	"github.com/spf13/cobra"
)

func LatestTagCommand() *cobra.Command {
	var githubApiUrl string
	var githubUploadUrl string
	var versionRange string
	cmd := &cobra.Command{
		Use:  "latest-tag [REPOSITORY] [GH_TOKEN]",
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			executeLatestTag(cmd, githubApiUrl, githubUploadUrl, versionRange, args)
		},
	}

	cmd.Flags().StringVarP(&githubApiUrl, "github-api-url", "a", "", "Github enterprise api url")
	cmd.Flags().StringVarP(&githubUploadUrl, "github-uploads-url", "u", "", "Github enterprise upload url")
	cmd.Flags().StringVarP(&versionRange, "version-range", "r", "", "Version range to select latest-tag from")

	return cmd
}

func executeLatestTag(cmd *cobra.Command, githubApiUrl, githubUploadUrl, versionRange string, args []string) {
	repository := args[0]
	githubToken := args[1]

	ctx := context.Background()

	client, err := enterprise.NewGithubClient(ctx, githubToken, githubApiUrl, githubUploadUrl)
	if err != nil {
		action.AssertNoError(cmd, err, "could not connect to github enterprise api: %s", err)
	}

	parts := strings.Split(repository, "/")
	owner := parts[0]
	repo := parts[1]

	refs, response, err := client.Git.ListMatchingRefs(ctx, owner, repo, &github.ReferenceListOptions{
		Ref: "tags",
	})
	if response != nil && response.StatusCode == http.StatusNotFound {
		cmd.Print("v0.0.0")
		return
	}
	action.AssertNoError(cmd, err, "could not list git refs: %s", err)
	expectedRange, err := semver.ParseRange(versionRange)
	action.AssertNoError(cmd, err, "could not create version range: %s", err)

	latest := semver.MustParse("0.0.0")
	for _, ref := range refs {
		version, err := semver.ParseTolerant(strings.Replace(*ref.Ref, "refs/tags/", "", 1))
		if err != nil {
			continue
		}
		if expectedRange(version) && version.GT(latest) {
			latest = version
		}
	}

	cmd.Printf("v%s", latest)
}
