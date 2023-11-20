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

type tag struct {
	major, minor, detail int
}

type Tags struct {
	tags []tag
}

func New() *Tags {
	return &Tags{
		tags: []tag{},
	}
}

func (t *Tags) Add(s string) error {
	if !strings.HasPrefix(s, "v") {
		return fmt.Errorf("tag %q doesn't start with 'v'", s)
	}
	parts := strings.Split(strings.TrimPrefix(s, "v"), ".")
	if len(parts) != 3 {
		return fmt.Errorf("tag %q doesn't have 3 parts, just %v", s, parts)
	}
	tg := tag{}
	var err error

	tg.major, err = strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("can't parse major number of tag %q (%v)", s, err)
	}
	tg.minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("can't parse minor number of tag %q (%v)", s, err)
	}
	tg.detail, err = strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("can't parse detail number of tag %q (%v)", s, err)
	}

	t.tags = append(t.tags, tg)
	return nil
}

func (t *Tags) HasTags() bool {
	return len(t.tags) > 0
}

func (t *Tags) Highest() string {
	if len(t.tags) == 0 {
		return anonymousTag
	}
	var high int
	for i := 1; i < len(t.tags); i++ {
		if t.tags[i].major > t.tags[high].major ||
			t.tags[i].minor > t.tags[high].minor ||
			t.tags[i].detail > t.tags[high].detail {
			high = i
		}
	}
	return fmt.Sprintf("v%d.%d.%d", t.tags[high].major, t.tags[high].minor, t.tags[high].detail)
}

func Next(s string) string {
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
