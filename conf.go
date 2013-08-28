package lab

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sync"
)

// Conf is special flag set that can be loaded via LoadConf in three passes.
var Conf = flag.NewFlagSet("config", flag.ContinueOnError)

// LoadConf loads the Conf flag set values in three passes on first invocation.
// The first pass will parse the arguments to populate the -conf flag.
// The second pass loads the file specified by the -conf flag.
// The third pass overrides all flags specifed by program arguments.
func LoadConf() {
	loadOnce.Do(loadConf)
}

var loadOnce sync.Once
var confFile = Conf.String("conf", "~/.golab/flags.conf", "flags config file")
var maxprocs = Conf.Int("maxprocs", 0, "GOMAXPROCS")

func loadConf() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		Conf.PrintDefaults()
	}
	err := Conf.Parse(os.Args[1:])
	if err != nil {
		os.Exit(2)
	}
	if *confFile != "" {
		loadConfFile()
	}
	err = Conf.Parse(os.Args[1:])
	if err != nil {
		log.Println("parsing config args", err)
	}
	maxprocs := *maxprocs
	if maxprocs < 0 {
		maxprocs = runtime.NumCPU()
	}
	if maxprocs > 0 {
		runtime.GOMAXPROCS(maxprocs)
	}
}

func loadConfFile() {
	path := *confFile
	if len(path) > 1 && path[0] == '~' && path[1] == '/' {
		usr, err := user.Current()
		if err != nil {
			log.Println("expanding home dir for config path", err)
			return
		}
		path = filepath.Join(usr.HomeDir, path[2:])
	}
	f, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println("opening config file", err)
		}
		return
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println("reading config file", err)
	}
	var toks []string
	var start int
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case ' ', '\t', '\n', '\r', '\f':
			if start < i {
				toks = append(toks, string(data[start:i]))
			}
			start = i + 1
		case '#':
			for j := i; j < len(data); j++ {
				if data[j] == '\n' {
					i = j
					start = i + 1
					break
				}
			}
		default:
		}
	}
	if start < len(data)-1 {
		toks = append(toks, string(data[start:]))
	}
	err = Conf.Parse(toks)
	if err != nil {
		log.Println("parsing config file", err)
	}
}
