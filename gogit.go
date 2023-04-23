package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	usageInfo = `
Usage: gogit pre-commit  # or: gogit stdfiles && gogit gotests
   or: gogit pre-push    # or: gogit allcommitted

`
)

func main() {
	if len(os.Args) != 2 {
		usage()
	}

	actions := map[string][]func() error{
		"pre-commit": {stdFiles, goTests},
		"stdfiles":   {stdFiles},
		"gotests":    {goTests},

		"pre-push":     {allCommitted, goVets},
		"allcommitted": {allCommitted},
		"govets":       {goVets},
	}
	funcs, ok := actions[os.Args[1]]
	if !ok {
		usage()
	}
	check(gotoGitTop())
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
		for _, l := range strings.Split(err.Error(), "\n") {
			fmt.Fprintf(os.Stdout, "[gogit] fatal: %v\n", l)
		}
		os.Exit(1)
	}
}

func gotoGitTop() error {
	lines, err := run("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("%v\nrun `git init` and retry", err)
	}
	if len(lines) != 1 {
		lines = append(lines, "need exactly 1 output to find the top level git folder")
		return errors.New(strings.Join(lines, "\n"))
	}
	fmt.Printf("[gogit] top level git folder: %q\n", lines[0])
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
			"  echo .git > .gitignore")
	}
	if _, err := os.Stat("go.mod"); err != nil {
		errs = append(errs,
			"`go.mod` not found, at a minimum run:",
			"  go mod init && go mod tidy")
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
		if strings.Contains(l, "working tree clean") {
			return nil
		}
	}
	return errors.New(
		strings.Join(
			append(lines,
				"not everything is commited, run:",
				"  git add $FILE(s)",
				"  git commit -m $MESSAGE"),
			"\n"))
}

func goVets() error {
	_, err := run("go", "vet", "./...")
	return err
}

func run(cmd ...string) ([]string, error) {
	fmt.Printf("[gogit] running %v\n", strings.Join(cmd, " "))
	b, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	lines := []string{}
	for _, l := range strings.Split(string(b), "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "[gogit] output:")
		for _, l := range lines {
			fmt.Fprintf(os.Stderr, "[gogit] %v\n", l)
		}
	}
	return lines, err
}
