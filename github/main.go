package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/Masterminds/semver"
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

// TagRepository creates a tag for a repository with the given commit sha and an optional list of files
// note: only the top level files in the directory will be uploaded, this function does not support subdirectories
func (m *Github) CreateRelease(ctx context.Context, owner, repo, tag, sha string, files Optional[*Directory]) error {
	client := m.getClient(ctx)

	rel, _, err := client.Repositories.CreateRelease(ctx, owner, repo, &github.RepositoryRelease{
		TagName:         &tag,
		TargetCommitish: &sha,
	})
	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	// if there are files to upload, upload them to the release
	fd, ok := files.Get()
	if ok {
		assets := os.TempDir()
		fd.Export(ctx, assets)

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
			}
		}
	}

	return nil
}

// BumpVersionWithPRTag returns a the next semantic version based on the presence of a PR tag
// i.e. if the PR has a tag of `major` and the current tag is `1.1.2` the next version will be `2.0.0`
// if the PR has a tag of `minor` and the current tag is `1.1.2` the next version will be `1.2.0`
// if the PR has a tag of `patch` and the current tag is `1.1.2` the next version will be `1.1.3`
func (m *Github) BumpVersionWithPRTag(ctx context.Context, owner, repo string, pr int) (string, error) {
	client := m.getClient(ctx)

	// get the tags from the pr
	prd, _, err := client.PullRequests.Get(ctx, owner, repo, pr)
	if err != nil {
		return "", fmt.Errorf("failed to get pull requests: %w", err)
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

	cv, _ := semver.NewVersion("v0.0.0")
	if len(versions) > 0 {
		cv = versions[len(versions)-1]
	}

	bump := ""

	// if there are multiple labels, get the highest one
	for _, l := range prd.Labels {
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
		&oauth2.Token{AccessToken: m.token},
	)

	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

// example: dagger call ftest-create-release --token=GITHUB_TOKEN --files=./testfiles
func (m *Github) FTestCreateRelease(ctx context.Context, token *Secret, files *Directory) error {
	m.token, _ = token.Plaintext(ctx)

	newVersion, err := m.BumpVersionWithPRTag(ctx, "jumppad-labs", "daggerverse", 1)
	if err != nil {
		return err
	}

	fmt.Println("new version", newVersion)

	return m.CreateRelease(ctx, "jumppad-labs", "daggerverse", newVersion, "dfb10f17e1821f7ded833206dc752fdabefe9aad", Opt[*Directory](files))
}

// example: dagger call ftest-bump-version-with-prtag --token=GITHUB_TOKEN
func (m *Github) FTestBumpVersionWithPRTag(ctx context.Context, token *Secret) (string, error) {
	m.token, _ = token.Plaintext(ctx)

	return m.BumpVersionWithPRTag(ctx, "jumppad-labs", "daggerverse", 1)
}
