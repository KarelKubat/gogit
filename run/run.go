package run

import (
	"os/exec"
	"strings"

	"github.com/KarelKubat/gogit/out"
)

var cache = make(map[string][]string)

// Exec runs a command and returns its output, unless it was run before, then its cached results are returned.
func Exec(title string, cmd []string) ([]string, error) {
	cli := strings.Join(cmd, " ")
	if cached, ok := cache[cli]; ok {
		return cached, nil
	}

	out.Title(title)
	out.Msg("running %v", cli)
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
	cache[cli] = lines
	return lines, err
}
