# Daggerverse module for GitHub

This module provides a set of classes and functions to interact with the GitHub API.


## CreateRelease

CreateRelease creates a new release in the specified repository. The tag used to create the release must
not already exist in the repository. This command will create a new release with the specified tag.

If an optional directory is provided, the command will upload all files in the directory to the release.
Note: Only files in the top level of the directory will be uploaded. Files in subdirectories will be ignored.

Parameters:
- `owner` (str): The owner of the repository.
- `repo` (str): The name of the repository.
- `tag` (str): The tag to create and to use for the release.
- `sha` (str): The commit SHA to create the release from.
- `files` (Directory): The list of files to upload and associate with the release.
- `token` (Secret, optional): The GitHub token to use for authentication, can also be set using `WithToken`.

Example:

```go
files := dag.Directory()

err := dag.Github().
  WithToken("<your token>").
  CreateRelease("jumppad-labs", "daggerverse", "0.1.2", "3fdsdfdf3434", files)

if err != nil {
    log.Fatal(err)
}
```

## NextVerstionFromAssociatedPRLabel

If there is an associated open PR for the commit SHA and that PR contains any labels "major", 
"minor", or "patch", this function will increment the largest semantic version tag in the 
repository by the corresponding label. If no labels are found an empty string will be returned.

For example: If the latest tag is "v1.2.3" and the PR contains the label "minor", the new tag will be "v1.3.0".

If there are multiple PRs associated with the commit, the highest label from any matching PR will be used.

Parameters:
- `owner` (str): The owner of the repository.
- `repo` (str): The name of the repository.
- `sha` (st): Commit SHA associated with a PR.
- `token` (Secret, optional): The GitHub token to use for authentication, can also be set using `WithToken`.

Returns:
- `string`: The new semantic version tag, empty if no pr labels are foun or the commit SHA.

Example:

```go
newTag, err := dag.Github().
  WithToken("<your token>").
  BumpVersionWithPRTag("jumppad-labs", "daggerverse", 123)

if err != nil {
    log.Fatal(err)
}
```

## GetOIDCToken

'GetOIDCToken' returns an OIDC token for the current GitHub actions run.

Parameters:
- `actionsRequestToken` (Secret): The GitHub Actions request token, provided by the environment variable `ACTIONS_ID_TOKEN_REQUEST_TOKEN`.
- `actionsTokenURL` (string): The GitHub Actions token URL, provided by the environment variable `ACTIONS_ID_TOKEN_REQUEST_URL`. 

Returns:
- `string`: The OIDC token.

Note: To use this function you need to set the `id-token` permission in your workflow file.

```yaml
jobs:

build:
  runs-on: ubuntu-latest
  permissions:
    id-token: write
    contents: read
```

Example:

```go
tkn, err := dag.Github().
  GetOIDCToken("<your token>", "<your url>")
```

## CommitFile
Creates a new commit in the given respoiory with the specified file changes.

Parameters:
- `owner` (str): The owner of the repository.
- `repo` (str): The name of the repository.
- `author` (str): The author of the commit.
- `email` (str): The email of the author.
- `message` (str): The commit message.
- `file` (File): The file to commit.
- `branch` (str, optional): The branch to commit to.

Example:

```go
dag.Github().CommitFile(
  ctx,
  "jumppad-labs", 
  "daggerverse", 
  "John Doe",
  "john@doe.com",
  "Updated file",
  file,
)
```

Returns:
- `string`: The SHA of the new commit.
- `error`: An error if the commit fails.

## WithToken

Sets the Github token to use for authentication.
