package ws

import (
	"runtime"
	"sync"
)

func MountAll(w *Ws, dirs []string) []error {
	errs := make([]error, len(dirs))
	if runtime.GOMAXPROCS(0) == 1 {
		for i, path := range dirs {
			_, errs[i] = w.Mount(path)
		}
	} else {
		var wg sync.WaitGroup
		wg.Add(len(dirs))
		mount := func(path string, err *error) {
			_, *err = w.Mount(path)
			wg.Done()
		}
		for i, path := range dirs {
			go mount(path, &errs[i])
		}
		wg.Wait()
	}
	for _, err := range errs {
		if err != nil {
			return errs
		}
	}
	return nil
}
