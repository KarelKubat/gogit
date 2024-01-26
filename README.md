# gogit

`gogit` is an all-in-one tool to make your Go projects more suitable for Github. It's my way of structuring my projects and I hope it'll be useful for you too.

<!-- toc -->
- [What it does](#what-it-does)
- [Installation](#installation)
- [Examples](#examples)
  - [<code>git commit</code> phase](#git-commit-phase)
    - [Fresh installation of git hooks](#fresh-installation-of-git-hooks)
    - [Some files are expected](#some-files-are-expected)
    - [After adding the files and committing locally](#after-adding-the-files-and-committing-locally)
    - [Suggestion to automatically refresh the Table of Contents](#suggestion-to-automatically-refresh-the-table-of-contents)
  - [<code>git push</code> phase](#git-push-phase)
    - [We need a local tag](#we-need-a-local-tag)
    - [Local tag should be pushed to remote](#local-tag-should-be-pushed-to-remote)
<!-- /toc -->

## What it does

`gogit` is meant to be installed as the pre-commit and pre-push hook in Git repositories, and is specifically aimed at Go projects.

In the pre-commit phase, it checks:

- That the usual files are present, a `README.md`, `LICENCE.MD`, a `.gitignore`, plus `go.mod` and `go.sum`,
- That `.go` files have corresponding `_test.go` tests (if not, dummy test frames can be created),
- That the tests pass,
- That `govet` is happy,
- The table of contents in `README.md` is refreshed. When `README.md` isn't set up for automatic table of contents management, actions are suggested to enable this.

In the pre-push phase, it checks all of the above, plus:

- That all local files are committed,
- That the repository is tagged (this requires a local tag, when none present, `v0.0.0` is suggested),
- That there is a remote repository,
- That next pushes to a remote repository use a "one-higher version" tag (e.g., `v3.14.15`, when the old tag is `v3.14.14`),
- If the version is non-zero (greater than `v0.0.0`) and if the package is not on pkg.go.dev, then a suggestion is made to register the package.

The purpose of `gogit` is to ensure some repository sanity, and to suggest steps to achieve that. `gogit` itself doesn't create or modify files, but it shows suggestions, and where possible, the right commands.

## Installation

Get the repository and just `go install gogit.go`. Then the first thing you'll want to do, is `cd` into a respository, run `gogit hooks`, and follow the instructions.

## Examples

The listings below are a few examples of what `gogit` suggests. The output on a terminal is colorized, which makes it nicely stand out, but can't be shown in this document.

### `git commit` phase

#### Fresh installation of git hooks

```plain
gogit hooks  # Let's ask `gogit` which hooks we need.

[gogit] finding top level git folder
[gogit] checking that .git/hooks are installed
[gogit] hook ".git/hooks/pre-commit" doesn't exist
[gogit] hook ".git/hooks/pre-push" doesn't exist
[gogit] suggestion(s):
  echo exec gogit pre-commit > .git/hooks/pre-commit
  chmod +x .git/hooks/pre-commit
  echo exec gogit pre-push > .git/hooks/pre-push
  chmod +x .git/hooks/pre-push
```

#### Some files are expected

```plain
git commit -a -m $MESSAGE  # The hooks are in place, `git commit` can now be used.

[gogit] checking that standard files are present
[gogit] `.gitignore` not found, create one and retry
[gogit] `go.mod` not found
[gogit] suggestion(s):
  echo .git > .gitignore
  go mod init
  go mod tidy
```

#### After adding the files and committing locally

```plain
git commit -a -m $MESSAGE

[gogit] finding top level git folder
[gogit] checking that .git/hooks are installed
[gogit] checking that standard files are present
[gogit] checking for go tests
[gogit] running go test ./...
[gogit] checking go vet on local packages
[gogit] running go vet ./...
On branch main
nothing to commit, working tree clean
```

#### Suggestion to automatically refresh the Table of Contents

```plain
git commit -a -m $MESSAGE

[gogit] (Not fatal) README.md has no Table of Contents section
[gogit] to have the TOC automatically updated, run:
[gogit] go install github.com/kubernetes-sigs/mdtoc@latest
[gogit] add   <!-- toc -->    to README.md
[gogit] add   <!-- /toc -->   to README.md
```

### `git push` phase

#### We need a local tag

```plain
git push

[gogit] checking for git tag validity
[gogit] checking local git tag
[gogit] running git tag
[gogit] local tag not found
[gogit] suggestion(s):
  git tag -a v0.0.0 -m v0.0.0
```

#### Local tag should be pushed to remote

```plain
git push

[gogit] local tag: "v2.0.10", remote tag: "v2.0.9"
[gogit] checking whether local repo is ahead of remote
[gogit] checking local status
[gogit] local tag v2.0.10 will need pushing to remote
[gogit] suggestion(s):
  git push origin v2.0.10
```

```go
func untabToc() error {
    out.Title("untabbing README.md")
    b, err := os.ReadFile("README.md")
    if err != nil {
        return err
    }
    untabbed := []string{}
    active := false
    changed := false
    for _, line := range strings.Split(string(b), "\n") {
        if len(line) == 0 {
            untabbed = append(untabbed, line)
            continue
        }
        if strings.HasPrefix(line, "```") {
            if line != "```" {
                active = !active
                untabbed = append(untabbed, line)
                continue
            }
        }
        if !active {
            untabbed = append(untabbed, line)
            continue
        }
        // Fix tabs following the first one.
        if strings.HasPrefix(line, "\t\t") {
            line = "\t    " + line[2:]
            changed = true
        }
        // Fix the first \t.
        for len(line) > 0 && line[0] == '\t' {
            line = "    " + line[1:]
            changed = true
        }
        untabbed = append(untabbed, line)
    }
    if !changed {
        return nil
    }

    if err := os.Rename("README.md", "README.md.org"); err != nil {
        return err
    }
    if err := os.WriteFile("README.md", []byte(strings.Join(untabbed, "\n")), 0644); err != nil {
        return fmt.Errorf("failed to ovwerwrite README.md: %v, original is in README.md.org", err)
    }

    return nil
}
```
