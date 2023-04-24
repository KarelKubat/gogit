# gogit

<!-- REMEMBER TO RUN:
  mdtoc --inplace README.md
-->
<!-- toc -->
- [What it does](#what-it-does)
- [Installation](#installation)
- [Examples](#examples)
  - [<code>git commit</code> phase](#git-commit-phase)
    - [Local repository commit](#local-repository-commit)
    - [Some files are expected](#some-files-are-expected)
    - [After adding the files and committing locally](#after-adding-the-files-and-committing-locally)
  - [<code>git push</code> phase](#git-push-phase)
    - [We need a tag](#we-need-a-tag)
    - [Success](#success)
<!-- /toc -->

`gogit` is an all-in-one tool to make your Go projects more suitable for Github. It's my way of structuring my projects and I hope it'll be useful for you too.

## What it does

`gogit` is meant to be installed as the pre-commit and pre-push hook in Git repositories, and is specifically aimed at Go projects.

In the pre-commit phase, it checks:

- That the usual files are present, a `README.md` and a `.gitignore`, plus `go.mod` and `go.sum`,
- That `.go` files have corresponding `_test.go` tests,
- That the tests pass,
- That `govet` is happy.

In the pre-push phase, it checks:

- That all local files are committed,
- That the repository is tagged (this requires the tag to be in a file `gittag.txt`),
- That there are remotes set up.

The purpose of `gogit` is to ensure some repository sanity, and to suggest steps to achieve that. `gogit` itself doesn't create or modify files, but it shows suggestions, and where possible, the right commands.

## Installation

Get the repository and just `go install gittag.go`.

## Examples

The listings below are a few examples of what `gogit` suggests. The output on a terminal is colorized, which makes it nicely stand out, but can't be shown in this document.

### `git commit` phase

#### Local repository commit

```plain
# Let's ask `gogit` which hooks we need.
gogit pre-commit
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
[gogit] git hook "pre-commit" succeeded, may your favorite git goat god smile on you
On branch main
nothing to commit, working tree clean
```

### `git push` phase

#### We need a tag

```plain
git push
[gogit] finding top level git folder
[gogit] running git rev-parse --show-toplevel
[gogit] top level git folder: "/Users/karelk/go/src/github.com/KarelKubat/gtpl"
[gogit] checking that .git/hooks are installed
[gogit] checking that standard files are present
[gogit] checking for go tests
[gogit] running go test ./...
[gogit] checking go vet on local packages
[gogit] running go vet ./...
[gogit] checking that everything is locally committed
[gogit] running git status
[gogit] checking for git tag validity
[gogit] "gittag.txt" not found, for a first tagging run:
[gogit] echo '# repository tag, update upon changes' > gittag.txt
[gogit] echo v0.0.0 >> gittag.txt
[gogit] git tag -a v0.0.0 -m $MESSAGE
[gogit] suggestion(s):
  echo '# repository tag, update upon changes' > gittag.txt
  echo v0.0.0 >> gittag.txt
  git tag -a v0.0.0 -m $MESSAGE
error: failed to push some refs to 'https://github.com/KarelKubat/gtpl.git'
```

#### Success

```plain
git push
[gogit] finding top level git folder
[gogit] running git rev-parse --show-toplevel
[gogit] top level git folder: "/Users/karelk/go/src/github.com/KarelKubat/gtpl"
[gogit] checking that .git/hooks are installed
[gogit] checking that standard files are present
[gogit] checking for go tests
[gogit] running go test ./...
[gogit] checking go vet on local packages
[gogit] running go vet ./...
[gogit] checking that everything is locally committed
[gogit] running git status
[gogit] checking for git tag validity
[gogit] running git tag
[gogit] checking for configured remote repositories
[gogit] running git remote
[gogit] "origin" is a remote repository
[gogit] git hook "pre-push" succeeded, may your favorite git goat god smile on you
Enumerating objects: 44, done.
Counting objects: 100% (44/44), done.
Delta compression using up to 10 threads
Compressing objects: 100% (42/42), done.
Writing objects: 100% (44/44), 39.25 KiB | 7.85 MiB/s, done.
Total 44 (delta 1), reused 0 (delta 0), pack-reused 0
remote: Resolving deltas: 100% (1/1), done.
To https://github.com/KarelKubat/gtpl.git
 + bd77803...903d297 main -> main
branch 'main' set up to track 'origin/main'.
```
