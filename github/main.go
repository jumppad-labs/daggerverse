package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Github struct {
	token string
}

// WithToken sets the GithHub token for any opeations that require it
func (m *Github) WithToken(token string) *Github {
	m.token = token

	return m
}

func (m *Github) TagRepository(ctx context.Context, owner, repo, tag, sha, message string) error {
	client := m.getClient(ctx)

	_, _, err := client.Git.CreateTag(ctx, owner, repo, &github.Tag{
		Tag:     &tag,
		Object:  &github.GitObject{SHA: &sha},
		Message: &message,
	})

	return err
}

// BumpVersionWithPRTag returns a the next semantic version based on the presence of a PR tag
// i.e. if the PR has a tag of `major` and the current tag is `1.1.2` the next version will be `2.0.0`
// if the PR has a tag of `minor` and the current tag is `1.1.2` the next version will be `1.2.0`
// if the PR has a tag of `patch` and the current tag is `1.1.2` the next version will be `1.1.3`
func (m *Github) BumpVersionWithPRTag(ctx context.Context, owner, repo string, pr int) (error, string) {
	client := m.getClient(ctx)

	// get the tags from the pr
	client.PullRequests.Get(ctx, owner, repo, pr)

	return "", err
}

func (m *Github) getClient(ctx context.Context) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: m.token},
	)

	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

func (m *Github) FTestCreateTag(ctx context.Context) error {
	token := os.Getenv("GITHUB_TOKEN")
	fmt.Println("token", token)

	return nil
}

func (m *Github) FTestBumpVersionWithPRTag(ctx context.Context) error {

	m.BumpVersionWithPRTag(ctx, "jumppad-labs", "daggerverse", 1)

	return nil
}
