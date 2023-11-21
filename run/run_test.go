package run

import (
	"testing"
)

func TestExec(t *testing.T) {
	for _, test := range []struct {
		cmd           []string
		wantErr       bool
		wantCacheSize int
	}{
		{
			cmd:           []string{"ls"},
			wantErr:       false,
			wantCacheSize: 1,
		},
	} {
		_, err := Exec("", test.cmd)
		gotErr := err != nil
		if gotErr != test.wantErr {
			t.Errorf("Exec(_,%v) = _,%q, goterr=%v, wanterr=%v", test.cmd, err, gotErr, test.wantErr)
		}
		if !gotErr {
			if gotCacheSize := len(cache); gotCacheSize != test.wantCacheSize {
				t.Errorf("len(cache) = %v, want %v", gotCacheSize, test.wantCacheSize)
			}
		}
	}
}
