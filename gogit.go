package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/KarelKubat/gogit/action"
	"github.com/KarelKubat/gogit/errs"
	"github.com/KarelKubat/gogit/out"
)

const (
	// Message when invoked without args
	usageInfo = `
Usage: gogit pre-commit  # or: gogit stdfiles && gogit gotests && gogit govets
   or: gogit pre-push    # or: gogit allcommitted && gogit gittag && go haveremote

`

	// `git status` output when nothing needs adding or committing
	statusOK = "working tree clean"

	// The git tag is tracked in this file
	gitTagFile = "gittag.txt"

	// Tag format as a regular expression, e.g. v1.0.12
	tagFormat = `v\d+\.\d+\.\d+`
)

func main() {
	if len(os.Args) != 2 {
		usage()
	}

	checks := map[string][]func() error{
		"pre-commit": {stdFiles, goTests, goVets},
		"stdfiles":   {stdFiles},
		"gotests":    {goTests},
		"govets":     {goVets},

		"pre-push":     {allCommitted, gitTag, haveRemote},
		"allcommitted": {allCommitted},
		"gittag":       {gitTag},
		"haveremote":   {haveRemote},
	}
	funcs, ok := checks[os.Args[1]]
	if !ok {
		usage()
	}
	check(gotoGitTop())
	check(hooksInstalled())
	for _, f := range funcs {
		check(f())
	}
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
	out.Title("finding top level git folder")
	lines, err := run("git", "rev-parse", "--show-toplevel")
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
	if _, err := os.Stat("README.md"); err != nil {
		errs.Add("`README.md` not found, create one and retry")
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
		_, err := run("go", "test", "./...")
		if err != nil {
			errs.Add(err.Error())
		}
	}
	return errs.Err()
}

func allCommitted() error {
	out.Title("checking that everything is locally committed")
	lines, err := run("git", "status")
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
	out.Title("checking go vet on local packages")
	_, err := run("go", "vet", "./...")
	return err
}

func gitTag() error {
	out.Title("checking for git tag validity")
	_, err := os.Stat(gitTagFile)
	if err != nil {
		return errors.New(
			strings.Join([]string{
				fmt.Sprintf("%q not found, for a first tagging run:", gitTagFile),
				action.Suggest("echo '# repository tag, update upon changes' > %v", gitTagFile),
				action.Suggest("echo v0.0.0 >> %v", gitTagFile),
				action.Suggest("git tag -a v0.0.0 -m $MESSAGE"),
			}, "\n"))
	}
	b, err := os.ReadFile(gitTagFile)
	if err != nil {
		return err
	}
	fileTag := ""
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "#") {
			continue
		}
		matched, err := regexp.MatchString(tagFormat, l)
		if err != nil {
			return fmt.Errorf("internal jam for regexp %q: %v", tagFormat, err)
		}
		if matched {
			fileTag = l
			break
		}
	}
	if fileTag == "" {
		return errors.New(
			strings.Join([]string{
				fmt.Sprintf("no tag found in %q", gitTagFile),
				"ensure that it holds a tag in the format vNR.NR.NR such as v1.0.12",
			}, "\n"))
	}
	lines, err := run("git tag")
	if err != nil {
		return err
	}
	lastTag := ""
	for _, tag := range lines {
		if tag == fileTag {
			return nil
		}
		lastTag = tag
	}
	errs := []string{
		fmt.Sprintf("file %q tags this release as %v, but the repository tags is %v", gitTagFile, fileTag, lastTag),
	}
	if lastTag < fileTag {
		errs = append(errs,
			"increase the repository tag, run:",
			action.Suggest("git tag -a %v -m $MESSAGE", fileTag),
		)
	} else {
		errs = append(errs,
			fmt.Sprintf("increase the tag in %q, run:", gitTagFile),
			action.Suggest("echo '# repository tag, update upon changes' > %v", gitTagFile),
			action.Suggest("echo %v >> %v", lastTag, gitTagFile))
	}
	return errors.New(strings.Join(errs, "\n"))
}

func haveRemote() error {
	out.Title("checking for configured remote repositories")
	lines, err := run("git remote")
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
			fmt.Sprintf("git remote add origin %v", githubURI))
	} else {
		suggestions = []string{"git remote add $REMOTE"}
	}
	errs.Add("no remote repository is configured")
	return errs.Add(suggestions...)
}

func run(cmd ...string) ([]string, error) {
	out.Msg(fmt.Sprintf("running %v\n", strings.Join(cmd, " ")))
	b, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	lines := []string{}
	for _, l := range strings.Split(string(b), "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	if err != nil {
		out.Error("output:")
		for _, l := range lines {
			out.Error(l)
		}
	}
	return lines, err
}
