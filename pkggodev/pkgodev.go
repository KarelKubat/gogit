package pkggodev

import (
	"fmt"
	"io"
	"net/http"
)

type Package struct {
	Name string
}

func New(name string) *Package {
	return &Package{
		Name: name,
	}
}

func (p *Package) URL() string {
	return "https://pkg.go.dev/" + p.Name
}

func (p *Package) HasPackage() (bool, error) {
	res, err := http.Get(p.URL())
	if err != nil {
		return false, fmt.Errorf("failed to query %q: %v", p.URL(), err)
	}
	_, err = io.ReadAll(res.Body)
	if err != nil {
		return false, fmt.Errorf("failed to fetch %q: %v", p.URL(), err)
	}
	res.Body.Close()
	return res.StatusCode == 200, nil
}
