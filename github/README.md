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

Example:

```go
files := dag.Directory()

err := dag.Github.CreateRelease("jumppad-labs", "daggerverse", "0.1.2", "3fdsdfdf3434", files)
if err != nil {
    log.Fatal(err)
}
```

## BumpVersionWithPRTag

If the referenced PR contains any labels "major", "minor", or "patch", this function will increment
the largest semantic version tag in the repository by the corresponding label. If no labels are found
an empty string will be returned.

Parameters:
- `owner` (str): The owner of the repository.
- `repo` (str): The name of the repository.
- `pr` (int): The pull request number.

Returns:
- `string`: The new semantic version tag, empty if no pr labels are found.

Example:

```go
newTag, err := dag.Github.BumpVersionWithPRTag("jumppad-labs", "daggerverse", 123)
if err != nil {
    log.Fatal(err)
}
```