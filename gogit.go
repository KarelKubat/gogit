package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/KarelKubat/gogit/action"
	"github.com/KarelKubat/gogit/errs"
	"github.com/KarelKubat/gogit/out"
	"github.com/KarelKubat/gogit/run"
)

const (
	// Message when invoked without args
	usageInfo = `
Usage:
  # check that we're in a git repository and suggest to install hooks
  gogit hooks

  # pre-commit checks
  gogit pre-commit  # or: gogit stdfiles && gogit gotests && gogit govets

  # pre-push checks, runs the above pre-commit checks first
  gogit pre-push    # or: gogit allcommitted && gogit haveremote && gogit gittag

`

	// `git status` output when nothing needs adding or committing
	statusOK = "working tree clean"

	// The git tag is tracked in this file
	gitTagFile = "gittag.txt"

	// Tag format as a regular expression, e.g. v1.0.12
	tagFormat = `v\d+\.\d+\.\d+`

	// Expected substring in output of `git ls-remote --tags` to find a tag
	remoteTagFormat = "refs/tags/"
)

func main() {
	if len(os.Args) != 2 {
		usage()
	}

	checks := map[string][]func() error{
		"hooks": {gotoGitTop, hooksInstalled},

		"pre-commit": {gotoGitTop, hooksInstalled, stdFiles, goTests, goVets},
		"stdfiles":   {gotoGitTop, hooksInstalled, stdFiles},
		"gotests":    {gotoGitTop, hooksInstalled, goTests},
		"govets":     {gotoGitTop, hooksInstalled, goVets},

		"pre-push":     {gotoGitTop, hooksInstalled, stdFiles, goTests, goVets, allCommitted, haveRemote, gitTag},
		"allcommitted": {gotoGitTop, hooksInstalled, allCommitted},
		"haveremote":   {gotoGitTop, hooksInstalled, haveRemote},
		"gittag":       {gotoGitTop, hooksInstalled, gitTag},
	}
	funcs, ok := checks[os.Args[1]]
	if !ok {
		usage()
	}
	for _, f := range funcs {
		check(f())
	}
	action.Output()
}

func usage() {
	fmt.Fprintf(os.Stderr, usageInfo)
	os.Exit(1)
}

func check(err error) {
	if err != nil {
		out.Error(err.Error())
		action.Output()
		os.Exit(1)
	}
}

func hooksInstalled() error {
	out.Title("checking that .git/hooks are installed")
	for _, hook := range []string{"pre-commit", "pre-push"} {
		path := fmt.Sprintf(".git/hooks/%v", hook)
		_, err := os.Stat(path)
		if err != nil {
			errs.Add(
				fmt.Sprintf("hook %q doesn't exist, run:", path),
				action.Suggest("echo exec gogit %v > %v", hook, path),
				action.Suggest("chmod +x %v", path))
		}
	}
	return errs.Err()
}

var gitTop string

func gotoGitTop() error {
	lines, err := run.Exec("finding top level git folder",
		[]string{"git", "rev-parse", "--show-toplevel"})
	if err != nil {
		return errs.Add(
			err.Error()+", try:",
			action.Suggest("git init"))
	}
	if len(lines) != 1 {
		lines = append(lines, "need exactly 1 output to find the top level git folder")
		return errs.Add(lines...)
	}
	gitTop = lines[0]
	out.Msg(fmt.Sprintf("top level git folder: %q\n", gitTop))
	if err := os.Chdir(gitTop); err != nil {
		return errs.Add(fmt.Sprintf("cannot chdir to top level git folder: %v", err))
	}
	return nil
}

func stdFiles() error {
	out.Title("checking that standard files are present")
	for _, md := range []string{"README.md", "LICENSE.md"} {
		if _, err := os.Stat(md); err != nil {
			errs.Add(fmt.Sprintf("file %v not found, create one and retry", md))
		}
	}
	if _, err := os.Stat(".gitignore"); err != nil {
		errs.Add(
			"`.gitignore` not found, create one and retry, at a minimum run:",
			action.Suggest("echo .git > .gitignore"))
	}
	if _, err := os.Stat("go.mod"); err != nil {
		errs.Add(
			"`go.mod` not found, at a minimum run:",
			action.Suggest("go mod init"),
			action.Suggest("go mod tidy"))
	}
	return errs.Err()
}

func goTests() error {
	out.Title("checking for go tests")
	srcs := map[string]struct{}{}
	tests := map[string]struct{}{}
	filepath.WalkDir(".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			errs.Add(err.Error())
			return err
		}
		switch {
		case strings.HasSuffix(p, "_test.go"):
			tests[p] = struct{}{}
		case strings.HasSuffix(p, ".go"):
			srcs[p] = struct{}{}
		}
		return nil
	})
	if errs.Err() != nil {
		return errs.Err()
	}
	testsFound := false
	for s := range srcs {
		wantTest := strings.Replace(s, ".go", "_test.go", 1)
		if _, ok := tests[wantTest]; !ok {
			errs.Add(fmt.Sprintf("go source %q lacks a test %q", s, wantTest))
		} else {
			testsFound = true
		}
	}
	if testsFound {
		_, err := run.Exec("running go tests",
			[]string{"go", "test", "./..."})
		if err != nil {
			errs.Add(err.Error())
		}
	}
	return errs.Err()
}

func allCommitted() error {
	lines, err := run.Exec("checking that everything is locally committed",
		[]string{"git", "status"})
	if err != nil {
		errs.Add(err.Error())
		return errs.Err()
	}
	for _, l := range lines {
		if strings.Contains(l, statusOK) {
			return nil
		}
	}
	errs.Add(lines...)
	return errs.Add(
		"not everything is commited, run:",
		action.Suggest("git status              # to check what's needed"),
		action.Suggest("git add $FILE(s)        # to add new files if needed"),
		action.Suggest("git commit -m $MESSAGE  # to locally commit"))
}

func goVets() error {
	_, err := run.Exec("checking go vet on local packages",
		[]string{"go", "vet", "./..."})
	return err
}

func gitTag() error {
	out.Title("checking for git tag validity")
	localTag, err := localGitTag()
	if err != nil {
		return err
	}
	remoteTag, err := remoteGitTag()
	if err != nil {
		return err
	}
	out.Msg(fmt.Sprintf("local tag: %v, remote tag: %v", localTag, remoteTag))
	if localTag <= remoteTag {
		return errors.New(
			strings.Join([]string{
				"the local tag must indicate a higher version than the remote one",
				"increase the local tag first, run:",
				action.Suggest("VERSION=v3.14.15  # example"),
				action.Suggest("git tag -a $VERSION -m $VERSION"),
			}, "\n"))
	}
	out.Msg(fmt.Sprintf("local tag %v will need pushing to remote, remember to run:", localTag))
	out.Msg(action.Suggest("git push origin %v", localTag))
	return nil
}

func localGitTag() (string, error) {
	lines, err := run.Exec("checking local git tag",
		[]string{"git", "tag"})
	if err != nil {
		return "", err
	}
	if len(lines) < 1 {
		return "", errors.New(strings.Join([]string{
			"local tag not found, for a first tagging, run:",
			action.Suggest("git tag -a v0.0.0 -m v0.0.0"),
		}, "\n"))
	}
	return lines[len(lines)-1], nil
}

func remoteGitTag() (string, error) {
	lines, err := run.Exec("checking remote git tag",
		[]string{"git", "ls-remote", "--tags"})
	if err != nil {
		return "", err
	}
	if len(lines) < 2 {
		return "", errors.New(strings.Join([]string{
			"remote tag not found, for a first tagging, run:",
			action.Suggest("git push origin v0.0.0"),
		}, "\n"))
	}
	re, err := regexp.Compile(tagFormat)
	if err != nil {
		return "", fmt.Errorf("internal jam for regexp %q: %v", tagFormat, err)
	}
	tag := ""
	for _, l := range lines {
		if !strings.Contains(l, remoteTagFormat) {
			continue
		}
		if t := re.FindString(l); t != "" {
			tag = t
		}
	}
	if tag == "" {
		errs := []string{
			"remote tags could not be parsed, for a first tagging, run:",
			action.Suggest("git tag -a v0.0.0 -m v0.0.0"),
			"output of command was:",
		}
		errs = append(errs, lines...)
		return "", errors.New(strings.Join(errs, "\n"))
	}
	return tag, nil
}

func haveRemote() error {
	lines, err := run.Exec("checking for remote repositories",
		[]string{"git", "remote"})
	if err != nil {
		return errs.Add(err.Error())
	}
	if len(lines) > 0 {
		for _, l := range lines {
			out.Msg(fmt.Sprintf("%q is a remote repository", l))
		}
		return nil
	}
	var suggestions []string
	repo := path.Base(gitTop)
	if strings.Contains(gitTop, "github.com") {
		_, after, found := strings.Cut(gitTop, "github.com")
		if !found || after == "" {
			return errs.Add(fmt.Sprintf("internal jam: failed to parse %q (after:%v, found:%v)", gitTop, after, found))
		}
		githubURI := fmt.Sprintf("https://github.com/%v", after)
		suggestions = append(suggestions,
			fmt.Sprintf("on github.com add the repository %v, and then:", repo),
			fmt.Sprintf("git remote add origin %v.git", githubURI))
	} else {
		suggestions = []string{"git remote add $REMOTE"}
	}
	errs.Add("no remote repository is configured")
	return errs.Add(suggestions...)
}
