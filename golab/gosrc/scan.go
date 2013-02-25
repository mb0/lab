package gosrc

import (
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/mb0/lab/ws"
)

func Scan(p *Pkg, r *ws.Res) {
	var srcs, tests, names, imprts []string
	flag := p.Flag ^ (HasSource | HasTest | HasXTest)
	r.Lock()
	path := r.Dir.Path
	for _, c := range r.Children {
		if c.Flag&(ws.FlagDir|FlagGo) != FlagGo {
			fp := filepath.Join(path, c.Name)
			if strings.HasSuffix(c.Name, "_test.go") {
				tests = append(tests, fp)
			} else {
				srcs = append(srcs, fp)
			}
		}
	}
	r.Unlock()
	if len(srcs) > 0 {
		flag |= HasSource
		parse(srcs, &names, &imprts)
	}
	if len(tests) > 0 {
		flag |= HasTest
		parse(tests, &names, &imprts)
	}
	name, xtest, err := checkname(names)
	if err != nil {
		fmt.Println(err)
	}
	p.Name = name
	if xtest {
		flag |= HasXTest
	}
	p.Flag = flag | Scanned
	fmt.Println("scanned pkg", p.Path)
}
func parse(files []string, names, imprts *[]string) {
	fset := token.NewFileSet()
	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments|parser.ImportsOnly)
		if err != nil {
			// log error
			continue
		}
		addname(names, f.Name.Name)
		for _, s := range f.Imports {
			addname(imprts, s.Path.Value)
		}
	}
}
func addname(names *[]string, name string) {
	for _, n := range *names {
		if n == name {
			return
		}
	}
	*names = append(*names, name)
}
func checkname(names []string) (name string, xtest bool, err error) {
	for _, n := range names {
		if strings.HasSuffix(n, "_test") {
			xtest = true
			n = n[:len(n)-5]
		}
		if name != "" && name != n {
			err = fmt.Errorf("multiple package names: %s %s", name, n)
			continue
		}
		name = n
	}
	return
}
