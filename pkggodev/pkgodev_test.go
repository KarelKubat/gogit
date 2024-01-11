package pkggodev

import (
	"testing"
)

func TestAll(t *testing.T) {
	for _, test := range []struct {
		name        string
		wantPresent bool
	}{
		{
			name:        "github.com/KarelKubat/gogit",
			wantPresent: true,
		},
		{
			name:        "github.com/KarelKubat/xyzzy-plugh",
			wantPresent: false,
		},
	} {
		p := New(test.name)
		present, err := p.HasPackage()
		if err != nil {
			t.Fatalf("%+v .HasPackage() = _,%q, need nil error", p, err.Error())
		}
		if present != test.wantPresent {
			t.Errorf("%+v .HasPackage() = %v,_, want %v", p, present, test.wantPresent)
		}
	}
}
