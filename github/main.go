package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/charmbracelet/log"
	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

type Github struct {
	Token string
}

// WithToken sets the GithHub token for any opeations that require it
func (m *Github) WithToken(token string) *Github {
	m.Token = token

	return m
}

// TagRepository creates a tag for a repository with the given commit sha and an optional list of files
// note: only the top level files in the directory will be uploaded, this function does not support subdirectories
func (m *Github) CreateRelease(
	ctx context.Context,
	owner,
	repo,
	tag,
	sha string,
	// +optional
	files *Directory,
) error {
	client := m.getClient(ctx)

	rel, _, err := client.Repositories.CreateRelease(ctx, owner, repo, &github.RepositoryRelease{
		TagName:         &tag,
		TargetCommitish: &sha,
	})

	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	log.Debug("Created release", "release", *rel.ID)

	// if there are files to upload, upload them to the release
	if files != nil {
		assets := os.TempDir()
		files.Export(ctx, assets)

		// get the files in the directory
		fs, err := os.ReadDir(assets)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}

		for _, f := range fs {
			if !f.IsDir() {
				iof, err := os.OpenFile(path.Join(assets, f.Name()), os.O_RDONLY, 0600)
				if err != nil {
					return fmt.Errorf("failed to open file: %w", err)
				}

				_, _, err = client.Repositories.UploadReleaseAsset(ctx, owner, repo, *rel.ID, &github.UploadOptions{Name: f.Name()}, iof)
				if err != nil {
					return fmt.Errorf("failed to upload file: %w", err)
				}

				log.Debug("Added file to release", "file", f.Name())
			}
		}
	}

	return nil
}

// NextVersionFromAssociatedPRLabel returns a the next semantic version based on the presence of a PR label
// for the given commit SHA.
// If there are multiple PRs associated with the commit, the highest label from any matching PR will be used
//
// i.e. if the SHA has an associated PR with a label of `major` and the current tag is `1.1.2` the next version will be `2.0.0`
// if the PR has a tag of `minor` and the current tag is `1.1.2` the next version will be `1.2.0`
// if the PR has a tag of `patch` and the current tag is `1.1.2` the next version will be `1.1.3`
func (m *Github) NextVersionFromAssociatedPRLabel(
	ctx context.Context,
	owner,
	repo,
	sha string,
) (string, error) {
	client := m.getClient(ctx)

	// find any associated PRs with the commit
	prs, _, err := client.PullRequests.ListPullRequestsWithCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get pull requests: %w", err)
	}

	// no PRS associated with this commit, return an empty string
	if len(prs) == 0 {
		log.Debug("No PRs associated with commit")
		return "", nil
	}

	// get the latest release tag
	tags, _, err := client.Repositories.ListTags(ctx, owner, repo, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get releases: %w", err)
	}

	versions := []*semver.Version{}
	for _, t := range tags {
		// check if the tag is a semver, if so add it to the list
		v, err := semver.NewVersion(*t.Name)
		if err == nil {
			versions = append(versions, v)
		}
	}

	// sort the list of semver tags
	sort.Sort(semver.Collection(versions))

	// create the new tag
	cv, _ := semver.NewVersion("v0.0.0")

	// if there were any tags, get the latest one
	if len(versions) > 0 {
		cv = versions[len(versions)-1]
	}

	bump := ""

	// check the PRs for labels
	for _, pr := range prs {
		log.Debug("Checking PR for labels", "pr", *pr.Number)

		// if there are multiple labels, get the highest one
		for _, l := range pr.Labels {
			switch *l.Name {
			case "major":
				bump = "major"
			case "minor":
				if bump != "major" {
					bump = "minor"
				}
			case "patch":
				if bump == "" {
					bump = "minor"
				}
			}
		}
	}

	log.Debug("Setting version increment", "bump", bump)

	switch bump {
	case "major":
		nv := cv.IncMajor()
		return nv.String(), nil
	case "minor":
		nv := cv.IncMinor()
		return nv.String(), nil
	case "patch":
		nv := cv.IncPatch()
		return nv.String(), nil
	}

	return "", nil
}

func (m *Github) getClient(ctx context.Context) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: m.Token},
	)

	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

// example: dagger call ftest-create-release --token=GITHUB_TOKEN --files=./testfiles
func (m *Github) FTestCreateRelease(
	ctx context.Context,
	token *Secret,
	// +optional
	files *Directory,
) error {
	// enable debug logging
	log.SetLevel(log.DebugLevel)

	m.Token, _ = token.Plaintext(ctx)

	v, err := m.NextVersionFromAssociatedPRLabel(ctx, "jumppad-labs", "daggerverse", "ee05014ca8f81bf9b2faae7f68d8c537bf7df577")
	if err != nil {
		return err
	}

	log.Debug("new version", "version", v)

	return m.CreateRelease(ctx, "jumppad-labs", "daggerverse", v, "ee05014ca8f81bf9b2faae7f68d8c537bf7df577", files)
}

// example: dagger call ftest-bump-version-with-prtag --token=GITHUB_TOKEN
func (m *Github) FTestBumpVersionWithPRTag(ctx context.Context, token *Secret) (string, error) {
	// enable debug logging
	log.SetLevel(log.DebugLevel)

	m.Token, _ = token.Plaintext(ctx)

	v, err := m.NextVersionFromAssociatedPRLabel(ctx, "jumppad-labs", "daggerverse", "ee05014ca8f81bf9b2faae7f68d8c537bf7df577")
	if err != nil {
		return v, err
	}

	log.Debug("new version", "version", v)

	return v, nil
}
