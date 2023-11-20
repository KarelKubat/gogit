// Package tags compares tags in the form v1.23.45.
package tags

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
	major, minor, detail int
}

type Tags struct {
	tags []Tag
}

func New() *Tags {
	return &Tags{
		tags: []Tag{},
	}
}

func (t *Tags) Add(s string) error {
	tg, err := ParseTag(s)
	if err != nil {
		return err
	}
	t.tags = append(t.tags, tg)
	return nil
}

func (t *Tags) HasTags() bool {
	return len(t.tags) > 0
}

func (t *Tags) Highest() Tag {
	if len(t.tags) == 0 {
		return Tag{0, 0, 0}
	}
	var high int
	for i := 1; i < len(t.tags); i++ {
		if t.tags[i].Greater(t.tags[high]) {
			high = i
		}
	}
	return t.tags[high]
}

func ParseTag(s string) (Tag, error) {
	tg := Tag{}
	if !strings.HasPrefix(s, "v") {
		return tg, fmt.Errorf("tag %q doesn't start with 'v'", s)
	}
	parts := strings.Split(strings.TrimPrefix(s, "v"), ".")
	if len(parts) != 3 {
		return tg, fmt.Errorf("tag %q doesn't have 3 parts, just %v", s, parts)
	}
	var err error

	tg.major, err = strconv.Atoi(parts[0])
	if err != nil {
		return tg, fmt.Errorf("can't parse major number of tag %q (%v)", s, err)
	}
	tg.minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return tg, fmt.Errorf("can't parse minor number of tag %q (%v)", s, err)
	}
	tg.detail, err = strconv.Atoi(parts[2])
	if err != nil {
		return tg, fmt.Errorf("can't parse detail number of tag %q (%v)", s, err)
	}
	return tg, nil
}

func (tg *Tag) Less(ot Tag) bool {
	switch {
	case tg.major < ot.major:
		return true
	case tg.major > ot.major:
		return false
	case tg.minor < ot.minor:
		return true
	case tg.minor > ot.minor:
		return false
	case tg.detail < ot.detail:
		return true
	case tg.detail > ot.detail:
		return false
	default:
		return false
	}
}

func (tg *Tag) Greater(ot Tag) bool {
	if tg.Less(ot) {
		return false
	}
	if tg.major == ot.major && tg.minor == ot.minor && tg.detail == ot.detail {
		return false
	}
	return true
}

func (tg Tag) String() string {
	return fmt.Sprintf("v%d.%d.%d", tg.major, tg.minor, tg.detail)
}

func (tg Tag) Next() string {
	nx := tg
	nx.detail++
	return nx.String()
}

func (tg Tag) IsZero() bool {
	return tg.major == 0 && tg.minor == 0 && tg.detail == 0
}

func NextTag(s string) string {
	if !TagRe.MatchString(s) {
		return anonymousTag
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return anonymousTag
	}
	lastNr, err := strconv.Atoi(parts[2])
	if err != nil {
		return anonymousTag
	}
	parts[2] = fmt.Sprintf("%d", lastNr+1)
	return strings.Join(parts, ".")
}
