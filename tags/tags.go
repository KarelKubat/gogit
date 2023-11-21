// Package tags compares tags in the form v1.23.45.
package tags

import (
	"github.com/KarelKubat/gogit/tag"
)

type Tags struct {
	tags []*tag.Tag
}

func New() *Tags {
	return &Tags{
		tags: []*tag.Tag{},
	}
}

func (t *Tags) Add(s string) error {
	tg, err := tag.New(s)
	if err != nil {
		return err
	}
	t.tags = append(t.tags, tg)
	return nil
}

func (t *Tags) HasTags() bool {
	return len(t.tags) > 0
}

func (t *Tags) Highest() *tag.Tag {
	if len(t.tags) == 0 {
		return tag.NewZero()
	}
	var high int
	for i := 1; i < len(t.tags); i++ {
		if t.tags[i].Greater(t.tags[high]) {
			high = i
		}
	}
	return t.tags[high]
}
