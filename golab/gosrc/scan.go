package gosrc

import (
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/mb0/lab/ws"
)

type Import struct {
	Path, Name string
	Id         ws.Id
}
type File struct {
	Name string
	Id   ws.Id
	Err  error
}

type Info struct {
	Files   []File
	Imports []Import
	Uses    []ws.Id
}

func (nfo *Info) Import(path string) *Import {
	for i, imprt := range nfo.Imports {
		if path == imprt.Path {
			return &nfo.Imports[i]
		}
	}
	return nil
}
func (nfo *Info) AddImport(path string) {
	if nfo.Import(path) == nil {
		nfo.Imports = append(nfo.Imports, Import{Path: path})
	}
}
func (nfo *Info) File(name string) *File {
	for i, file := range nfo.Files {
		if name == file.Name {
			return &nfo.Files[i]
		}
	}
	return nil
}
func (nfo *Info) AddFile(name string, id ws.Id) {
	if nfo.File(name) == nil {
		nfo.Files = append(nfo.Files, File{Name: name, Id: id})
	}
}
func (nfo *Info) AddUse(id ws.Id) {
	for _, i := range nfo.Uses {
		if i == id {
			return
		}
	}
	nfo.Uses = append(nfo.Uses, id)
}

func Scan(p *Pkg, r *ws.Res) error {
	fmt.Println("scan pkg", p.Path)
	p.Flag ^= HasSource | HasTest | HasXTest
	src, test := getinfo(p, r)
	var err error
	if src != nil {
		p.Flag |= HasSource
		p.Name, err = parse(p, src, "")
		if p.Src != nil {
			src.Uses = p.Src.Uses
		}
	}
	if test != nil {
		p.Flag |= HasTest
		p.Name, err = parse(p, test, p.Name)
	}
	p.Src, p.Test = src, test
	p.Flag |= Scanned
	return err
}
func getinfo(p *Pkg, r *ws.Res) (src, test *Info) {
	r.Lock()
	defer r.Unlock()
	if r.Dir == nil {
		return
	}
	p.Dir = r.Dir.Path
	for _, c := range r.Children {
		if c.Flag&(ws.FlagDir|FlagGo) == FlagGo {
			if strings.HasSuffix(c.Name, "_test.go") {
				if test == nil {
					test = &Info{}
				}
				test.AddFile(c.Name, c.Id)
			} else {
				if src == nil {
					src = &Info{}
				}
				src.AddFile(c.Name, c.Id)
			}
		}
	}
	return
}
func parse(p *Pkg, info *Info, name string) (string, error) {
	fset := token.NewFileSet()
	var lasterr error
	for _, file := range info.Files {
		path := filepath.Join(p.Dir, file.Name)
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.ImportsOnly)
		if err != nil {
			lasterr, file.Err = err, err
			continue
		}
		var ok, xtest bool
		if name, ok, xtest = checkname(f.Name.Name, name); !ok {
			lasterr = fmt.Errorf("package name err: %s %s", f.Name.Name, name)
			file.Err = lasterr
			continue
		} else if xtest {
			p.Flag |= HasXTest
		}
		for _, s := range f.Imports {
			info.AddImport(s.Path.Value[1 : len(s.Path.Value)-1])
		}
	}
	return name, lasterr
}
func checkname(name string, now string) (string, bool, bool) {
	var xtest bool
	if strings.HasSuffix(name, "_test") {
		xtest = true
		name = name[:len(name)-5]
	}
	if now == "" || name == now {
		return name, name != "", xtest
	}
	return now, false, xtest
}
