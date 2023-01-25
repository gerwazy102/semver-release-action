package release

import (
	"context"
	"fmt"
	"strings"

	"github.com/gerwazy102/semver-release-action/internal/pkg/action"
	"github.com/gerwazy102/semver-release-action/internal/pkg/enterprise"
	"github.com/google/go-github/v45/github"
	"github.com/spf13/cobra"
)

const releaseTypeNone = "none"
const releaseTypeRelease = "release"
const releaseTypeTag = "tag"

type repository struct {
	owner string
	name  string
	token string
}

type releaseDetails struct {
	version string
	target  string
}

func Command() *cobra.Command {
	var releaseType string
	var githubApiUrl string
	var githubUploadUrl string

	cmd := &cobra.Command{
		Use:  "release [REPOSITORY] [TARGET_COMMITISH] [VERSION] [GH_TOKEN]",
		Args: cobra.ExactArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			execute(cmd, releaseType, githubApiUrl, githubUploadUrl, args)
		},
	}

	cmd.Flags().StringVarP(&releaseType, "strategy", "s", releaseTypeRelease, "Release strategy")
	cmd.Flags().StringVarP(&githubApiUrl, "github-api-url", "a", "", "Github enterprise api url")
	cmd.Flags().StringVarP(&githubUploadUrl, "github-uploads-url", "u", "", "Github enterprise upload url")

	return cmd
}

func execute(cmd *cobra.Command, releaseType, githubApiUrl, githubUploadUrl string, args []string) {
	parts := strings.Split(args[0], "/")
	repo := repository{
		owner: parts[0],
		name:  parts[1],
		token: args[3],
	}

	release := releaseDetails{
		version: args[2],
		target:  args[1],
	}

	ctx := context.Background()

	client, err := enterprise.NewGithubClient(ctx, repo.token, githubApiUrl, githubUploadUrl)
	if err != nil {
		action.AssertNoError(cmd, err, "could not connect to github enterprise api: %s", err)
	}

	switch releaseType {
	case releaseTypeNone:
		return
	case releaseTypeRelease:
		if err := createGithubRelease(ctx, client, repo, release); err != nil {
			action.AssertNoError(cmd, err, "could not create GitHub release: %s", err)
		}
		return
	case releaseTypeTag:
		if err := createLightweightTag(ctx, client, repo, release); err != nil {
			action.AssertNoError(cmd, err, "could not create lightweight tag: %s", err)
		}
		return
	default:
		action.Fail(cmd, "unknown release strategy: %s", releaseType)
	}
}

func createLightweightTag(ctx context.Context, client *github.Client, repo repository, release releaseDetails) error {
	_, _, err := client.Git.CreateRef(ctx, repo.owner, repo.name, &github.Reference{
		Ref: github.String(fmt.Sprintf("refs/tags/%s", release.version)),
		Object: &github.GitObject{
			SHA: &release.target,
		},
	})

	return err
}

func createGithubRelease(ctx context.Context, client *github.Client, repo repository, release releaseDetails) error {
	_, _, err := client.Repositories.CreateRelease(ctx, repo.owner, repo.name, &github.RepositoryRelease{
		Name:            &release.version,
		TagName:         &release.version,
		TargetCommitish: &release.target,
		Draft:           github.Bool(false),
		Prerelease:      github.Bool(false),
	})

	return err
}
