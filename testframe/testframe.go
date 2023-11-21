package testframe

import (
	"fmt"
	"os"
	"strings"
)

const (
	testFrame = `package %v

import (
	"testing"
)

func TestAll(t *testing.T) {
	// TODO: Add tests
}
`
)

func Make(src string) error {
	// General checks
	if strings.HasSuffix(src, "_test.go") {
		return fmt.Errorf("%v looks like a test file already", src)
	}
	_, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("can't stat %v: %v", src, err)
	}
	tst := strings.TrimSuffix(src, ".go")
	tst += "_test.go"
	_, err = os.Stat(tst)
	if err == nil {
		return fmt.Errorf("test file %v already exists, won't overwrite", tst)
	}

	// Find the package name
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read %v: %v", src, err)
	}
	pkgame, err := packageName(content, src)
	if err != nil {
		return err
	}
	if err = os.WriteFile(tst, []byte(fmt.Sprintf(testFrame, pkgame)), 0644); err != nil {
		return fmt.Errorf("failed to write %v: %v", tst, err)
	}

	return nil
}

func packageName(content []byte, fname string) (string, error) {
	pkgame := ""
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "package") {
			parts := strings.Split(line, " ")
			if len(parts) != 2 {
				return "", fmt.Errorf("package statement %q must have just 2 parts", line)
			}
			pkgame = parts[1]
			break
		}
	}
	if pkgame == "" {
		return "", fmt.Errorf("failed to extract package name from %v", fname)
	}
	return pkgame, nil
}
