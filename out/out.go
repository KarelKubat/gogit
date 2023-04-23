package out

import (
	"fmt"
	"strings"

	"github.com/mitchellh/colorstring"
)

func out(col, msg string) {
	if msg != "" {
		colorstring.Printf(fmt.Sprintf("[gogit] [%v]%v\n", col, msg))
	}
}

func Error(ss ...string) {
	for _, s := range ss {
		for _, l := range strings.Split(s, "\n") {
			out("red", l)
		}
	}
}

func Title(ss ...string) {
	for _, s := range ss {
		for _, l := range strings.Split(s, "\n") {
			out("yellow", l)
		}
	}
}

func Msg(ss ...string) {
	for _, s := range ss {
		for _, l := range strings.Split(s, "\n") {
			out("green", l)
		}
	}
}
