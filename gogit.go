package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/KarelKubat/gogit/action"
	"github.com/KarelKubat/gogit/errs"
	"github.com/KarelKubat/gogit/out"
	"github.com/KarelKubat/gogit/run"
	"github.com/KarelKubat/gogit/tag"
	"github.com/KarelKubat/gogit/tags"
	"github.com/KarelKubat/gogit/testframe"
)

const (
	// Message when invoked without args
	usageInfo = `
Usage:
  # check that we're in a git repository and suggest to install hooks
  gogit hooks

  # pre-commit checks
  gogit pre-commit  # or: gogit stdfiles && gogit gotests && gogit govets && gogit mdtoc

  # pre-push checks, runs the above pre-commit checks first
  gogit pre-push    # or: gogit allcommitted && gogit haveremote && gogit gittag

  # create a test frame for .go sources
  gogit make-test-frame a.go sub/b.go  # creates a_test.go and sub/b_test.go

`

	// `git status` output when nothing needs adding or committing
	statusOK = "working tree clean"

	// Expected substring in output of `git ls-remote --tags` to find a tag
	remoteTagFormat = "refs/tags/"

	// Tags in README.md to refresh the ToC
	tocStart = "<!-- toc -->"
	tocEnd   = "<!-- /toc -->"
)

func main() {
	// `gogit make-test-frame $GO_SRC` is a special case.
	if len(os.Args) >= 2 && os.Args[1] == "make-test-frame" {
		if len(os.Args) == 2 {
			usage()
		}
		for _, s := range os.Args[2:] {
			check(testframe.Make(s))
		}
		os.Exit(0)
	}

	// All other invocations have just one argument: the action to perform.
	if len(os.Args) != 2 {
		usage()
	}

	checks := map[string][]func() error{
		"hooks": {gotoGitTop, hooksInstalled},

		"pre-commit": {gotoGitTop, hooksInstalled, stdFiles, goTests, goVets, mdToc},
		"stdfiles":   {gotoGitTop, hooksInstalled, stdFiles},
		"gotests":    {gotoGitTop, hooksInstalled, goTests},
		"govets":     {gotoGitTop, hooksInstalled, goVets},
		"mdtoc":      {mdToc},

		"pre-push":     {gotoGitTop, hooksInstalled, stdFiles, goTests, goVets, mdToc, allCommitted, haveRemote, gitTag},
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
	suggestFrames := []string{}
	for s := range srcs {
		wantTest := strings.Replace(s, ".go", "_test.go", 1)
		if _, ok := tests[wantTest]; !ok {
			errs.Add(fmt.Sprintf("go source %q lacks a test %q", s, wantTest))
			suggestFrames = append(suggestFrames, s)
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
	if len(suggestFrames) > 0 {
		errs.Add("at a minimum run:")
		for _, frame := range suggestFrames {
			action.Suggest("gogit make-test-frame %v", frame)
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

func mdToc() error {
	b, err := ioutil.ReadFile("README.md")
	if err != nil {
		return err
	}
	nTags := 0
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, tocStart) || strings.HasPrefix(l, tocEnd) {
			nTags++
		}
	}
	switch nTags {
	case 0:
		out.Error(strings.Join([]string{
			"(Not fatal) README.md has no Table of Contents section",
			"to have the TOC automatically updated, run:",
			action.Suggest("go install github.com/kubernetes-sigs/mdtoc@latest"),
			action.Suggest("add   %v    to README.md (at first column)", tocStart),
			action.Suggest("add   %v   to README.md (at first column)", tocEnd),
		}, "\n"))
		return nil
	case 2:
		_, err := run.Exec("refreshing README.md TOC",
			[]string{"mdtoc", "--inplace", "README.md"})
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("README.md must contain exactly one tag `%v` and one tag `%v`, found: %v", tocStart, tocEnd, nTags)
	}
	return nil // to satisfy the signature
}

func gitTag() error {
	out.Title("checking git tags")
	localTag, err := localGitTag()
	if err != nil {
		return err
	}
	remoteTag, err := remoteGitTag()
	if err != nil {
		return err
	}
	out.Msg(fmt.Sprintf("local tag: %q, remote tag: %q", localTag, remoteTag))
	if !remoteTag.IsZero() && !localTag.Greater(remoteTag) {
		nextTag := remoteTag.Next()
		return errors.New(
			strings.Join([]string{
				"the local tag should indicate a higher version than the remote one",
				"increase the local tag first, run:",
				action.Suggest("# to increase the tag ID and push:"),
				action.Suggest("git tag -a %v -m %v  # or increase major/minor numbers", nextTag, nextTag),
				action.Suggest("git push"),
				action.Suggest("git push origin %v", nextTag),
				"alternatively, to stay on the same tag number, run:",
				action.Suggest("# to stay on the same tag ID, run:"),
				action.Suggest("git push --no_verify"),
			}, "\n"))
	}
	out.Msg(fmt.Sprintf("local tag %v will need pushing to remote, remember to run:", localTag))
	out.Msg(action.Suggest("git push origin %v", localTag))
	return nil
}

func localGitTag() (*tag.Tag, error) {
	lines, err := run.Exec("checking local git tag",
		[]string{"git", "tag"})
	if err != nil {
		return nil, err
	}
	if len(lines) < 1 {
		return nil, errors.New(strings.Join([]string{
			"local tag not found, for a first tagging, run:",
			action.Suggest("git tag -a v0.0.0 -m v0.0.0"),
		}, "\n"))
	}
	tgs := tags.New()
	for _, l := range lines {
		if err = tgs.Add(l); err != nil {
			return nil, errors.New(strings.Join([]string{
				err.Error(),
				"manually correct using:",
				action.Suggest("git tag -d %v", l),
			}, "\n"))
		}
	}
	return tgs.Highest(), nil
}

func remoteGitTag() (*tag.Tag, error) {
	lines, err := run.Exec("checking remote git tag",
		[]string{"git", "ls-remote", "--tags"})
	if err != nil {
		return nil, err
	}
	if len(lines) < 2 {
		return nil, nil
	}
	tgs := tags.New()
	for _, l := range lines {
		if !strings.Contains(l, remoteTagFormat) {
			continue
		}
		if t := tag.TagRe.FindString(l); t != "" {
			if err := tgs.Add(t); err != nil {
				errs := []string{
					"remote tags could not be parsed, for a first tagging, run:",
					action.Suggest("git tag -a v0.0.0 -m v0.0.0"),
					"output of command was:",
				}
				errs = append(errs, lines...)
				return nil, errors.New(strings.Join(errs, "\n"))
			}
		}
	}
	if !tgs.HasTags() {
		return nil, nil
	}
	return tgs.Highest(), nil
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
