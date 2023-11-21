package tag

import (
	"testing"
)

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
		if gotNext := Next(test.s); gotNext != test.wantNext {
			t.Errorf("Next(%q) = %q, want %q", test.s, gotNext, test.wantNext)
		}
	}
}

func TestLessGreaterEqual(t *testing.T) {
	for _, test := range []struct {
		tg          *Tag
		ot          *Tag
		wantLess    bool
		wantGreater bool
		wantEqual   bool
	}{
		{
			tg:          &Tag{Major: 1, Minor: 2, Detail: 3},
			ot:          &Tag{Major: 1, Minor: 2, Detail: 3},
			wantLess:    false,
			wantGreater: false,
			wantEqual:   true,
		},
		{
			tg:          &Tag{Major: 1, Minor: 2, Detail: 3},
			ot:          &Tag{Major: 1, Minor: 2, Detail: 4},
			wantLess:    true,
			wantGreater: false,
			wantEqual:   false,
		},
		{
			tg:          &Tag{Major: 1, Minor: 2, Detail: 4},
			ot:          &Tag{Major: 1, Minor: 2, Detail: 3},
			wantLess:    false,
			wantGreater: true,
			wantEqual:   false,
		},
		{
			tg:          &Tag{Major: 1, Minor: 2, Detail: 3},
			ot:          &Tag{Major: 10, Minor: 2, Detail: 4},
			wantLess:    true,
			wantGreater: false,
			wantEqual:   false,
		},
		{
			tg:          &Tag{Major: 10, Minor: 2, Detail: 4},
			ot:          &Tag{Major: 1, Minor: 2, Detail: 3},
			wantLess:    false,
			wantGreater: true,
			wantEqual:   false,
		},
		{
			tg:          &Tag{Major: 2, Minor: 0, Detail: 0},
			ot:          &Tag{Major: 1, Minor: 0, Detail: 10},
			wantLess:    false,
			wantGreater: true,
			wantEqual:   false,
		},
	} {
		if gotLess := test.tg.Less(test.ot); gotLess != test.wantLess {
			t.Errorf("%+v .Less(%+v): got %v, want %v", test.tg, test.ot, gotLess, test.wantLess)
		}
		if gotGreater := test.tg.Greater(test.ot); gotGreater != test.wantGreater {
			t.Errorf("%+v .Greater(%+v): got %v, want %v", test.tg, test.ot, gotGreater, test.wantGreater)

		}
		if gotEqual := test.tg.Equal(test.ot); gotEqual != test.wantEqual {
			t.Errorf("%+v .Equal(%+v): got %v, want %v", test.tg, test.ot, gotEqual, test.wantEqual)

		}
	}
}

func TestIsZero(t *testing.T) {
	for _, test := range []struct {
		tg         *Tag
		wantIsZero bool
	}{
		{
			tg:         nil,
			wantIsZero: true,
		},
		{
			tg:         &Tag{Major: 0, Minor: 0, Detail: 0},
			wantIsZero: true,
		},
		{
			tg:         &Tag{Major: 1, Minor: 0, Detail: 0},
			wantIsZero: false,
		},
		{
			tg:         &Tag{Major: 0, Minor: 1, Detail: 0},
			wantIsZero: false,
		},
		{
			tg:         &Tag{Major: 0, Minor: 0, Detail: 1},
			wantIsZero: false,
		},
	} {
		if gotZero := test.tg.IsZero(); gotZero != test.wantIsZero {
			t.Errorf("%+v .IsZero() = %v, want %v", test.tg, gotZero, test.wantIsZero)
		}
	}
}
