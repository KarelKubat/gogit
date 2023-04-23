package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mitchellh/colorstring"
)

const (
	// Message when invoked without args
	usageInfo = `
Usage: gogit pre-commit  # or: gogit stdfiles && gogit gotests
   or: gogit pre-push    # or: gogit allcommitted && gogit govets && gogit gittag

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

	actions := map[string][]func() error{
		"pre-commit": {stdFiles, goTests},
		"stdfiles":   {stdFiles},
		"gotests":    {goTests},

		"pre-push":     {allCommitted, goVets, gitTag},
		"allcommitted": {allCommitted},
		"govets":       {goVets},
		"gittag":       {gitTag},
	}
	funcs, ok := actions[os.Args[1]]
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

var suggestions []string

func suggest(f string, args ...interface{}) string {
	s := fmt.Sprintf(f, args...)
	suggestions = append(suggestions, "  "+s)
	return s
}

func check(err error) {
	if err != nil {
		for _, l := range strings.Split(err.Error(), "\n") {
			colorstring.Fprintf(os.Stdout, "[gogit] [red]fatal: %v\n", l)
		}
		if len(suggestions) > 0 {
			colorstring.Fprintf(os.Stdout, "[gogit] [yellow]suggestion(s):\n")
			for _, s := range suggestions {
				colorstring.Fprintf(os.Stdout, "[yellow]%v\n", s)
			}
		}
		os.Exit(1)
	}
}

func hooksInstalled() error {
	errs := []string{}
	for _, hook := range []string{"pre-commit", "pre-push"} {
		path := fmt.Sprintf(".git/hooks/%v", hook)
		_, err := os.Stat(path)
		if err != nil {
			errs = append(errs,
				fmt.Sprintf("hook %q doesn't exist, run:", path),
				suggest("echo exec gogit %v > %v", hook, path),
				suggest("chmod +x %v", path))
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func gotoGitTop() error {
	lines, err := run("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return errors.New(strings.Join([]string{
			err.Error() + ", try:",
			suggest("git init"),
		}, "\n"))
	}
	if len(lines) != 1 {
		lines = append(lines, "need exactly 1 output to find the top level git folder")
		return errors.New(strings.Join(lines, "\n"))
	}
	colorstring.Printf("[gogit] [green]top level git folder: %q\n", lines[0])
	if err := os.Chdir(lines[0]); err != nil {
		return fmt.Errorf("cannot chdir to top level git folder: %v", err)
	}
	return nil
}

func stdFiles() error {
	errs := []string{}
	if _, err := os.Stat("README.md"); err != nil {
		errs = append(errs, "`README.md` not found, create one and retry")
	}
	if _, err := os.Stat(".gitignore"); err != nil {
		errs = append(errs,
			"`.gitignore` not found, create one and retry",
			"at a minimum run:",
			suggest("echo .git > .gitignore"))
	}
	if _, err := os.Stat("go.mod"); err != nil {
		errs = append(errs,
			"`go.mod` not found, at a minimum run:",
			suggest("go mod init"),
			suggest("go mod tidy"))
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func goTests() error {
	errs := []string{}
	srcs := map[string]struct{}{}
	tests := map[string]struct{}{}
	filepath.WalkDir(".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			errs = append(errs, err.Error())
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
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	testsFound := false
	for s := range srcs {
		wantTest := strings.Replace(s, ".go", "_test.go", 1)
		if _, ok := tests[wantTest]; !ok {
			errs = append(errs, fmt.Sprintf("go source %q lacks a test %q", s, wantTest))
		} else {
			testsFound = true
		}
	}
	if testsFound {
		_, err := run("go", "test", "./...")
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func allCommitted() error {
	lines, err := run("git", "status")
	if err != nil {
		return err
	}
	for _, l := range lines {
		if strings.Contains(l, statusOK) {
			return nil
		}
	}
	return errors.New(
		strings.Join(
			append(lines,
				"not everything is commited, run:",
				suggest("git add $FILE(s)"),
				suggest("git commit -m $MESSAGE")),
			"\n"))
}

func goVets() error {
	_, err := run("go", "vet", "./...")
	return err
}

func gitTag() error {
	_, err := os.Stat(gitTagFile)
	if err != nil {
		return errors.New(
			strings.Join([]string{
				fmt.Sprintf("%q not found, for a first tagging run:", gitTagFile),
				suggest("echo '# repository tag, update upon changes' > %v", gitTagFile),
				suggest("echo v0.0.0 >> %v", gitTagFile),
				suggest("git tag -a v0.0.0 -m $MESSAGE"),
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
			suggest("git tag -a %v -m $MESSAGE", fileTag),
		)
	} else {
		errs = append(errs,
			fmt.Sprintf("increase the tag in %q, run:", gitTagFile),
			suggest("echo '# repository tag, update upon changes' > %v", gitTagFile),
			suggest("echo %v >> %v", lastTag, gitTagFile))
	}
	return errors.New(strings.Join(errs, "\n"))
}

func run(cmd ...string) ([]string, error) {
	colorstring.Printf("[gogit] [green]running %v\n", strings.Join(cmd, " "))
	b, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	lines := []string{}
	for _, l := range strings.Split(string(b), "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	if err != nil {
		colorstring.Fprintln(os.Stderr, "[gogit] output:")
		for _, l := range lines {
			colorstring.Fprintf(os.Stderr, "[gogit] %v\n", l)
		}
	}
	return lines, err
}
