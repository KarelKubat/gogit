package errs

import (
	"errors"
	"strings"
)

var errs []string

func Add(strs ...string) error {
	for _, s := range strs {
		if s != "" {
			errs = append(errs, s)
		}
	}
	return Err()
}

func Err() error {
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "\n"))
}
