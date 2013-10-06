// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package find

import (
	"fmt"
	"github.com/mb0/lab/ws"
	"os"
	"testing"
)

var mockdata = map[string][]string{
	"/r/1/a":       {"x.go", "y.go", "z.go"},
	"/r/1/b":       {"x.js", "y.js", "z.js"},
	"/r/1/c":       {"x.cx", "y.cy", "z.cz"},
	"/r/2/z.go":    {},
	"/r/3/foo":     {"a", "b", "c"},
	"/r/3/foo/bar": {"a", "b", "c"},
}

var queryTests = []struct {
	query  string
	result []string
}{
	{"x.go", []string{"x.go"}},
	{"*.go", []string{"x.go", "y.go", "z.go", "z.go/"}},
	{"x.*", []string{"x.go", "x.js", "x.cx"}},
	{"*.c*", []string{"x.cx", "y.cy", "z.cz"}},
	{"c/*x", []string{"x.cx"}},
	{"z.go$", []string{"z.go"}},
	{"z.go/", []string{"z.go/"}},
	{"c/*x", []string{"x.cx"}},
	{"3/*/a", []string{"a"}},
	{"3/**/a", []string{"a", "a"}},
	{"foo/**", []string{"bar/", "a", "b", "c", "a", "b", "c"}},
	{"foo/**/", []string{"bar/"}},
	{"foo/**$", []string{"a", "b", "c", "a", "b", "c"}},
	{"foo/**a", []string{"a", "a"}},
	{"foo/b**", []string{"bar/", "b", "b"}},
	{"foo/**a*", []string{"bar/", "a", "a"}},
	{"foo/*a**", []string{"bar/", "a", "a", "b", "c"}},
}

func TestFind(t *testing.T) {
	w := ws.New(ws.Config{Backend: MockFS(mockdata)})
	w.Mount("/r")
	for _, test := range queryTests {
		found, err := Find(w, test.query)
		if err == nil {
			err = namesEqual(found, test.result)
		}
		if err != nil {
			t.Errorf("query %q: %s", test.query, err)
		}
	}
}

func namesEqual(a []*ws.Res, b []string) error {
	if len(a) != len(b) {
		for _, r := range a {
			fmt.Println(r.Path())
		}
		return fmt.Errorf("length %d does not match %d", len(a), len(b))
	}
	for i, r := range a {
		name := b[i]
		if r.Dir != nil {
			name = name[0 : len(name)-1]
		}
		if name != r.Name {
			return fmt.Errorf("expected %s got %s", name, r.Name)
		}
	}
	return nil
}

func MockFS(flat map[string][]string) *ws.Backend {
	m := make(map[string]*MockFile)
	var mkdir func(path string) *MockFile

	mkdir = func(path string) *MockFile {
		path = ws.Filesystem.Clean(path)
		f, ok := m[path]
		if !ok {
			f = &MockFile{Path: path}
			m[path] = f
			dir, file := ws.Filesystem.Split(path)
			if dir == "" && file != path {
				dir = file
			}
			p := mkdir(dir)
			p.Children = append(p.Children, f)
		}
		f.Dir = true
		return f
	}
	for path, files := range flat {
		d := mkdir(path)
		for _, f := range files {
			d.Children = append(d.Children, &MockFile{Path: path + "/" + f})
		}
		m[path] = d
	}
	isDir := func(path string) (bool, error) {
		path = ws.Filesystem.Clean(path)
		f := m[path]
		if f == nil {
			return false, fmt.Errorf("no such file")
		}
		return f.Dir, nil
	}
	readDir := func(path string) ([]os.FileInfo, error) {
		path = ws.Filesystem.Clean(path)
		f := m[path]
		if f == nil {
			return nil, fmt.Errorf("no such file")
		}
		list := make([]os.FileInfo, len(f.Children))
		for i, c := range f.Children {
			list[i] = c
		}
		return list, nil
	}
	return &ws.Backend{
		IsDir:     isDir,
		ReadDir:   readDir,
		Seperator: ws.Filesystem.Seperator,
		Clean:     ws.Filesystem.Clean,
		Split:     ws.Filesystem.Split,
	}
}

type MockFile struct {
	os.FileInfo
	Path     string
	Dir      bool
	Children []*MockFile
}

func (f *MockFile) IsDir() bool {
	return f.Dir
}

func (f *MockFile) Name() string {
	_, name := ws.Filesystem.Split(f.Path)
	return name
}
