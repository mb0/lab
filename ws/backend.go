package ws

import (
	"os"
	"path/filepath"
)

type Backend struct {
	IsDir     func(path string) (bool, error)
	ReadDir   func(path string) ([]os.FileInfo, error)
	Seperator string
	Clean     func(path string) string
	Split     func(path string) (dir, file string)
}

var Filesystem = Backend{
	IsDir: func(path string) (bool, error) {
		fi, err := os.Stat(path)
		if err != nil {
			return false, err
		}
		return fi.IsDir(), nil
	},
	ReadDir: func(path string) ([]os.FileInfo, error) {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return f.Readdir(-1)
	},
	Seperator: string(os.PathSeparator),
	Clean:     filepath.Clean,
	Split:     filepath.Split,
}
