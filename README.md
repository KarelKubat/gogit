# gogit

<!-- REMEMBER TO RUN:
  mdtoc --inplace README.md
-->
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

`gogit` is an all-in-one tool to make your Go projects more suitable for Github. It's my way of structuring my projects and I hope it'll be useful for you too.

## What it does

`gogit` is meant to be installed as the pre-commit and pre-push hook in Git repositories, and is specifically aimed at Go projects.

In the pre-commit phase, it checks:

- That the usual files are present, a `README.md`, `LICENCE.MD`, a `.gitignore`, plus `go.mod` and `go.sum`,
- That `.go` files have corresponding `_test.go` tests (if not, dummy test frames can be created),
- That the tests pass,
- That `govet` is happy,
- When `README.md` isn't set up for automatic table of contents management, actions are suggested to enable this.

In the pre-push phase, it checks all of the above, plus:

- That all local files are committed,
- That the repository is tagged (this requires a local tag, when none present, `v0.0.0` is suggested),
- That next pushes to a remote repository use a "next version" tag (e.g., `v3.14.15`, when the old tag is `v3.14.14`),
- That there are remotes set up.

The purpose of `gogit` is to ensure some repository sanity, and to suggest steps to achieve that. `gogit` itself doesn't create or modify files, but it shows suggestions, and where possible, the right commands.

## Installation

Get the repository and just `go install gittag.go`. Then the first thing you'll want to do, is `cd` into a respository, run `gogit hooks`, and follow the instructions.

## Examples

The listings below are a few examples of what `gogit` suggests. The output on a terminal is colorized, which makes it nicely stand out, but can't be shown in this document.

### `git commit` phase

#### Fresh installation of git hooks

```plain
# Let's ask `gogit` which hooks we need.
gogit hooks
[gogit] finding top level git folder
[gogit] running git rev-parse --show-toplevel
[gogit] top level git folder: "/Users/karelk/go/src/github.com/KarelKubat/gtpl"
[gogit] checking that .git/hooks are installed
[gogit] hook ".git/hooks/pre-commit" doesn't exist, run:
[gogit] echo exec gogit pre-commit > .git/hooks/pre-commit
[gogit] chmod +x .git/hooks/pre-commit
[gogit] hook ".git/hooks/pre-push" doesn't exist, run:
[gogit] echo exec gogit pre-push > .git/hooks/pre-push
[gogit] chmod +x .git/hooks/pre-push
[gogit] suggestion(s):
  echo exec gogit pre-commit > .git/hooks/pre-commit
  chmod +x .git/hooks/pre-commit
  echo exec gogit pre-push > .git/hooks/pre-push
  chmod +x .git/hooks/pre-push
```

#### Some files are expected

```plain
# The hooks are in place, `git commit` can now be used.
git commit
[gogit] finding top level git folder
[gogit] running git rev-parse --show-toplevel
[gogit] top level git folder: "/Users/karelk/go/src/github.com/KarelKubat/gtpl"
[gogit] checking that .git/hooks are installed
[gogit] checking that standard files are present
[gogit] `.gitignore` not found, create one and retry, at a minimum run:
[gogit] echo .git > .gitignore
[gogit] `go.mod` not found, at a minimum run:
[gogit] go mod init
[gogit] go mod tidy
[gogit] suggestion(s):
  echo .git > .gitignore
  go mod init
  go mod tidy
```

#### After adding the files and committing locally

```plain
git commit -a -m 'conversion to gogit'
[gogit] finding top level git folder
[gogit] running git rev-parse --show-toplevel
[gogit] top level git folder: "/Users/karelk/go/src/github.com/KarelKubat/gtpl"
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
...
[gogit] (Not fatal) README.md has no Table of Contents section
[gogit] to have the TOC automatically updated, run:
[gogit] go install github.com/kubernetes-sigs/mdtoc@latest
[gogit] add   <!-- toc -->    to README.md
[gogit] add   <!-- /toc -->   to README.md
...
```

### `git push` phase

#### We need a local tag

```plain
git push
...
[gogit] checking for git tag validity
[gogit] checking local git tag
[gogit] running git tag
[gogit] local tag not found, for a first tagging, run:
[gogit] git tag -a v0.0.0 -m v0.0.0
[gogit] suggestion(s):
  git tag -a v0.0.0 -m v0.0.0
```

#### Local tag should be pushed to remote

```plain
...
[gogit] running git tag
[gogit] checking remote git tag
[gogit] running git ls-remote --tags
[gogit] remote tag not found, for a first tagging, run:
[gogit] git push origin v0.0.0
[gogit] suggestion(s):
  git push origin v0.0.0
  ```
