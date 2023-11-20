package tags

import (
	"testing"
)

func TestAll(t *testing.T) {
	tgs := New()
	for _, test := range []struct {
		s        string
		wantErr  bool
		wantHigh string
	}{
		{
			s:        "v1.0.0",
			wantErr:  false,
			wantHigh: "v1.0.0",
		},
		{
			s:        "v1.0.1",
			wantErr:  false,
			wantHigh: "v1.0.1",
		},
		{
			s:        "v1.0.2",
			wantErr:  false,
			wantHigh: "v1.0.2",
		},
		{
			s:        "v1.0.10",
			wantErr:  false,
			wantHigh: "v1.0.10",
		},
		{
			s:        "v1.1.2",
			wantErr:  false,
			wantHigh: "v1.1.2",
		},
	} {
		err := tgs.Add(test.s)
		gotErr := err != nil
		if gotErr != test.wantErr {
			t.Errorf("Add(%q): error = %v, got error = %v, want error = %v", test.s, err, gotErr, test.wantErr)
			continue
		}

		wantHasTags := len(tgs.tags) > 0
		if gotHasTags := tgs.HasTags(); gotHasTags != wantHasTags {
			t.Errorf("HasTags() = %v, want %v", gotHasTags, wantHasTags)
		}

		if gotHigh := tgs.Highest().String(); gotHigh != test.wantHigh {
			t.Errorf("Highest() = %q, want %q", gotHigh, test.wantHigh)
		}
	}
}

func TestNext(t *testing.T) {
	for _, test := range []struct {
		s        string
		wantNext string
	}{
		{
			s:        "v0.0.0",
			wantNext: "v0.0.1",
		},
		{
			s:        "v0.0.1",
			wantNext: "v0.0.2",
		},
		{
			s:        "v0.0.9",
			wantNext: "v0.0.10",
		},
		{
			s:        "v12.34.99",
			wantNext: "v12.34.100",
		},
		{
			s:        "",
			wantNext: "$TAG",
		},
	} {
		if gotNext := NextTag(test.s); gotNext != test.wantNext {
			t.Errorf("Next(%q) = %q, want %q", test.s, gotNext, test.wantNext)
		}
	}
}

func TestLessGreater(t *testing.T) {
	for _, test := range []struct {
		tg          Tag
		ot          Tag
		wantLess    bool
		wantGreater bool
	}{
		{
			tg:          Tag{1, 2, 3},
			ot:          Tag{1, 2, 3},
			wantLess:    false,
			wantGreater: false,
		},
		{
			tg:          Tag{1, 2, 3},
			ot:          Tag{1, 2, 4},
			wantLess:    true,
			wantGreater: false,
		},
		{
			tg:          Tag{1, 2, 4},
			ot:          Tag{1, 2, 3},
			wantLess:    false,
			wantGreater: true,
		},
		{
			tg:          Tag{1, 2, 3},
			ot:          Tag{10, 2, 4},
			wantLess:    true,
			wantGreater: false,
		},
		{
			tg:          Tag{10, 2, 4},
			ot:          Tag{1, 2, 3},
			wantLess:    false,
			wantGreater: true,
		},
		{
			tg:          Tag{2, 0, 0},
			ot:          Tag{1, 0, 10},
			wantLess:    false,
			wantGreater: true,
		},
	} {
		if gotLess := test.tg.Less(test.ot); gotLess != test.wantLess {
			t.Errorf("%+v .Less(%+v): got %v, want %v", test.tg, test.ot, gotLess, test.wantLess)
		}
	}
}
