// Package tag represents a git tag in the form v12.34.56. It is parsed into 3 numbers to make tags
// comparable.
package tag

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	TagFormat    = `v\d+\.\d+\.\d+` // ex. v1.23.45
	anonymousTag = "$TAG"           // placeholder
)

var TagRe = regexp.MustCompile(TagFormat)

type Tag struct {
	Major, Minor, Detail int
}

func New(s string) (*Tag, error) {
	tg := &Tag{}
	if !strings.HasPrefix(s, "v") {
		return tg, fmt.Errorf("tag %q doesn't start with 'v'", s)
	}
	parts := strings.Split(strings.TrimPrefix(s, "v"), ".")
	if len(parts) != 3 {
		return tg, fmt.Errorf("tag %q doesn't have 3 parts, just %v", s, parts)
	}
	var err error

	tg.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return tg, fmt.Errorf("can't parse major number of tag %q (%v)", s, err)
	}
	tg.Minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return tg, fmt.Errorf("can't parse minor number of tag %q (%v)", s, err)
	}
	tg.Detail, err = strconv.Atoi(parts[2])
	if err != nil {
		return tg, fmt.Errorf("can't parse detail number of tag %q (%v)", s, err)
	}
	return tg, nil
}

func NewZero() *Tag {
	return &Tag{
		Major:  0,
		Minor:  0,
		Detail: 0,
	}
}

func (tg *Tag) Less(ot *Tag) bool {
	switch {
	case tg.Major < ot.Major:
		return true
	case tg.Major > ot.Major:
		return false
	case tg.Minor < ot.Minor:
		return true
	case tg.Minor > ot.Minor:
		return false
	case tg.Detail < ot.Detail:
		return true
	case tg.Detail > ot.Detail:
		return false
	default:
		return false
	}
}

func (tg *Tag) Equal(ot *Tag) bool {
	if ot == nil {
		return false
	}
	return tg.Major == ot.Major && tg.Minor == ot.Minor && tg.Detail == ot.Detail
}

func (tg *Tag) Greater(ot *Tag) bool {
	if tg.Less(ot) || tg.Equal(ot) {
		return false
	}
	return true
}

func (tg *Tag) String() string {
	return fmt.Sprintf("v%d.%d.%d", tg.Major, tg.Minor, tg.Detail)
}

func (tg *Tag) Next() *Tag {
	nx := &Tag{
		Major:  tg.Major,
		Minor:  tg.Minor,
		Detail: tg.Detail,
	}
	nx.Detail++
	return nx
}

func (tg *Tag) IsZero() bool {
	return tg == nil || (tg.Major == 0 && tg.Minor == 0 && tg.Detail == 0)
}

func Next(s string) string {
	tg, err := New(s)
	if err != nil {
		return anonymousTag
	}
	return tg.Next().String()
}
