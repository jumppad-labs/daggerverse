package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/charmbracelet/log"
	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

type Github struct {
	Token *Secret
}

// WithToken sets the GithHub token for any opeations that require it
func (m *Github) WithToken(token *Secret) *Github {
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
	name string,
	// +optional
	files *Directory,
) error {
	client, err := m.getClient(ctx)
	if err != nil {
		return err
	}

	if name == "" {
		name = tag
	}

	rel, _, err := client.Repositories.CreateRelease(ctx, owner, repo, &github.RepositoryRelease{
		Name:            &name,
		TagName:         &tag,
		TargetCommitish: &sha,
	})
	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	tagMessage := "Create new release"
	_, _, err = client.Git.CreateTag(ctx, owner, repo, &github.Tag{
		Tag:     &tag,
		SHA:     &sha,
		Message: &tagMessage,
		Object:  &github.GitObject{SHA: &sha, Type: github.String("commit")},
	})
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
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
// If there are multiple PRs associated with the commit, the label from the latest PR will be used.
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
	client, err := m.getClient(ctx)
	if err != nil {
		return "", err
	}

	// find any associated PRs with the commit
	prs := []*github.PullRequest{}
	page := 0

	// loop through and get all prs associated with the commit, list might be paged
	for {
		p, resp, err := client.PullRequests.ListPullRequestsWithCommit(ctx, owner, repo, sha, &github.ListOptions{Page: page})
		if err != nil {
			return "", fmt.Errorf("failed to get pull requests: %w", err)
		}

		prs = append(prs, p...)

		if resp.NextPage == 0 {
			break
		}
	}

	// no PRS associated with this commit, return an empty string
	if len(prs) == 0 {
		log.Debug("No PRs associated with commit")
		return "", nil
	}

	versions := []*semver.Version{}
	page = 0

	for {
		// get the latest release tag
		tags, resp, err := client.Repositories.ListTags(ctx, owner, repo, &github.ListOptions{Page: page})
		if err != nil {
			return "", fmt.Errorf("failed to get releases: %w", err)
		}

		for _, t := range tags {
			// check if the tag is a semver, if so add it to the list
			v, err := semver.NewVersion(*t.Name)
			if err == nil {
				versions = append(versions, v)
			}
		}

		page = resp.NextPage
		if page == 0 {
			break
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
	maxID := 0

	// check the PRs for labels
	for _, pr := range prs {
		log.Debug("Checking PR for labels", "pr", *pr.Number, "labels", pr.Labels)

		// only check the latest closed PR
		if pr.Number != nil && *pr.Number > maxID {
			// if there are multiple labels, get the highest one
			lab := ""
			for _, l := range pr.Labels {
				switch *l.Name {
				case "major":
					lab = "major"
				case "minor":
					if lab != "major" {
						lab = "minor"
					}
				case "patch":
					if lab == "" {
						lab = "patch"
					}
				}
			}

			maxID = *pr.Number
			bump = lab
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

// GetOIDCToken returns an OpenID Connect (OIDC) token for the current run in GitHubActions
// When a actions run has the `id-token: write` permission, it can request an OIDC token for the current run
// the parameters actionsRequestToken and actionsTokenURL are provided by the GitHubActions environment
// variables `ACTIONS_ID_TOKEN_REQUEST_TOKEN` and `ACTIONS_ID_TOKEN_REQUEST_URL`.
//
// example actions config to enable OIDC tokens:
// jobs:
//
//	build:
//	  runs-on: ubuntu-latest
//	  permissions:
//	    id-token: write
//	    contents: read
func (m *Github) GetOIDCToken(ctx context.Context, actionsRequestToken *Secret, actionsTokenURL string, audience string) (string, error) {
	if audience != "" {
		actionsTokenURL = fmt.Sprintf("%s&audience=%s", actionsTokenURL, audience)
	}
	rq, err := http.NewRequest(http.MethodGet, actionsTokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create request: %w", err)
	}

	tkn, _ := actionsRequestToken.Plaintext(ctx)

	// add the bearer token for the request
	rq.Header.Add("Authorization", fmt.Sprintf("bearer %s", tkn))

	// make the request
	resp, err := http.DefaultClient.Do(rq)
	if err != nil {
		return "", fmt.Errorf("unable to request token: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// parse the response
	data := map[string]interface{}{}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response body: %w", err)
	}

	json.Unmarshal(body, &data)
	gitHubJWT := data["value"].(string)
	return gitHubJWT, nil
}

// CommitFile commits a file to a repository at the given path
func (m *Github) CommitFile(
	ctx context.Context,
	owner,
	repo,
	commiterName,
	commiterEmail,
	commitPath,
	message string,
	file *File,
	// +optional
	branch string,
) (string, error) {
	c, err := m.getClient(ctx)
	if err != nil {
		return "", err
	}

	var commitBranch *string

	if branch != "" {
		commitBranch = &branch
	}

	outPath := path.Join(os.TempDir(), "file")
	_, err = file.Export(ctx, outPath)
	if err != nil {
		return "", fmt.Errorf("failed to export file: %w", err)
	}

	data, _ := os.ReadFile(outPath)

	// get the sha of the existing file
	var sha *string

	f, _, _, err := c.Repositories.GetContents(ctx, owner, repo, commitPath, &github.RepositoryContentGetOptions{Ref: branch})
	if err != nil {
		// check if the error is a 404 error
		if !strings.Contains(err.Error(), "404") {
			return "", fmt.Errorf("failed to get file: %w", err)
		}
	}

	if f != nil {
		sha = f.SHA
	}

	cm, _, err := c.Repositories.UpdateFile(ctx, owner, repo, commitPath,
		&github.RepositoryContentFileOptions{
			Content: data,
			Branch:  commitBranch,
			Message: &message,
			SHA:     sha,
			Committer: &github.CommitAuthor{
				Name:  &commiterName,
				Email: &commiterEmail,
			},
		})
	if err != nil {
		return "", fmt.Errorf("failed to update file: %w", err)
	}

	return *cm.Commit.SHA, nil
}

func (m *Github) getClient(ctx context.Context) (*github.Client, error) {
	if m.Token == nil {
		log.Error("GitHub token not set")
		return nil, fmt.Errorf("GitHub token not set, please use the WithToken function to set the token")
	}

	tkn, err := m.Token.Plaintext(ctx)
	if err != nil {
		return nil, err
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tkn},
	)

	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc), nil
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

	m.Token = token

	v, err := m.NextVersionFromAssociatedPRLabel(ctx, "jumppad-labs", "daggerverse", "6976eb3f392256c384e87094853853f90c64ca68")
	if err != nil {
		return err
	}

	log.Debug("new version", "version", v)

	return m.CreateRelease(ctx, "jumppad-labs", "daggerverse", v, "6976eb3f392256c384e87094853853f90c64ca68", "", files)
}

// example: dagger call ftest-bump-version-with-prtag --token=GITHUB_TOKEN
func (m *Github) FTestBumpVersionWithPRTag(ctx context.Context, token *Secret) (string, error) {
	// enable debug logging
	log.SetLevel(log.DebugLevel)

	m.Token = token

	v, err := m.NextVersionFromAssociatedPRLabel(ctx, "jumppad-labs", "jumppad", "18e75c8517831bc29f5ce25528787c239ec670c1")
	if err != nil {
		return v, err
	}

	log.Debug("new version", "version", v)

	return v, nil
}

func (m *Github) FTestCommitFile(ctx context.Context, token *Secret) (string, error) {
	// enable debug logging
	log.SetLevel(log.DebugLevel)

	m.Token = token

	f := dag.Directory().WithNewFile("test.txt", time.Now().String()).File("test.txt")

	return m.CommitFile(ctx,
		"jumppad-labs",
		"daggerverse",
		"jumppad",
		"hello@jumppad.dev",
		"test.txt",
		"commit message",
		f,
		"",
	)
}
