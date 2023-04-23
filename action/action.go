package action

import (
	"fmt"
	"os"

	"github.com/mitchellh/colorstring"
)

var suggestions []string

func Suggest(f string, args ...interface{}) string {
	s := fmt.Sprintf(f, args...)
	suggestions = append(suggestions, "  "+s)
	return s
}

func Output() {
	if len(suggestions) > 0 {
		colorstring.Fprintf(os.Stdout, "[gogit] [yellow]suggestion(s):\n")
		for _, s := range suggestions {
			colorstring.Fprintf(os.Stdout, "[yellow]%v\n", s)
		}
	}
}
