package action

import (
	"testing"
)

func TestSuggest(t *testing.T) {
	for i := 0; i < 10; i++ {
		Suggest("this is suggestion %v", i)
	}
	if slen := len(suggestions); slen != 10 {
		t.Errorf("len(suggestions) = %v, want 10", slen)
	}
}
