package testframe

import (
	"strings"
	"testing"
)

func TestPackageName(t *testing.T) {
	for _, test := range []struct {
		content     string
		wantErr     string
		wantPackage string
	}{
		{
			content:     "package foo",
			wantErr:     "",
			wantPackage: "foo",
		},
		{
			content:     "package foo bar",
			wantErr:     "must have just 2 parts",
			wantPackage: "",
		},
		{
			content:     "foo bar",
			wantErr:     "failed to extract package name",
			wantPackage: "",
		},
	} {
		gotPackage, err := packageName([]byte(test.content), "src.go")
		switch {
		case err == nil && test.wantErr != "":
			t.Errorf("packagename(%q,_) = nil, want error with %q", test.content, test.wantErr)
		case err != nil && test.wantErr == "":
			t.Errorf("packagename(%q,_) = _,%q, want nil error", test.content, err.Error())
		case err != nil && test.wantErr != "" && !strings.Contains(err.Error(), test.wantErr):
			t.Errorf("packagename(%q,_) = _,%q, want error with %q", test.content, err.Error(), test.wantErr)
		case err == nil && test.wantErr == "" && gotPackage != test.wantPackage:
			t.Errorf("packagename(%q,_) = %q,_, want package %q", test.content, gotPackage, test.wantPackage)
		}
	}
}
