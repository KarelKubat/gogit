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

		if gotHigh := tgs.Highest(); gotHigh != test.wantHigh {
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
		if gotNext := Next(test.s); gotNext != test.wantNext {
			t.Errorf("Next(%q) = %q, want %q", test.s, gotNext, test.wantNext)
		}
	}
}
