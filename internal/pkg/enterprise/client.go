package enterprise

import (
	"context"
	"strings"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

func NewGithubClient(ctx context.Context, token, githubApiUrl, githubUploadUrl string) (*github.Client, error) {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	client := github.NewClient(oauth2.NewClient(ctx, tokenSource))
	if githubApiUrl != "" {
		if githubUploadUrl == "" {
			githubUploadUrl = strings.Replace(githubApiUrl, "api", "uploads", 1)
		}
		var err error
		client, err = github.NewEnterpriseClient(githubApiUrl, githubUploadUrl, oauth2.NewClient(ctx, tokenSource))
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}
