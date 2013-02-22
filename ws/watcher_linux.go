package ws // derived from exp/inotify

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

var (
	createMask uint32 = syscall.IN_CREATE | syscall.IN_MOVED_TO
	modifyMask uint32 = syscall.IN_CLOSE_WRITE
	deleteMask uint32 = syscall.IN_MOVED_FROM | syscall.IN_DELETE | syscall.IN_DELETE_SELF
	flagMask          = (createMask | modifyMask | deleteMask) ^ syscall.IN_DELETE_SELF
)

type inotify struct {
	sync.Mutex
	watchfd int
	watches map[Id]int32
	ids     map[int32]Id
	done    chan bool
	ctrler  Controller
}

func NewInotify(ctrler Controller) (Watcher, error) {
	watchfd, errno := syscall.InotifyInit()
	if watchfd == -1 {
		return nil, os.NewSyscallError("inotify_init", errno)
	}
	w := &inotify{
		watchfd: watchfd,
		watches: make(map[Id]int32),
		ids:     make(map[int32]Id),
		done:    make(chan bool, 1),
		ctrler:  ctrler,
	}
	go w.readEvents()
	return w, nil
}

func (w *inotify) Watch(r *Res) error {
	w.Lock()
	defer w.Unlock()
	if _, found := w.watches[r.Id]; found {
		return fmt.Errorf("duplicate watch")
	}
	flags := flagMask
	if r.Flag&FlagMount != 0 {
		flags |= syscall.IN_DELETE_SELF
	}
	return w.add(r.Id, r.Path(), flags)
}
func (w *inotify) add(id Id, path string, flags uint32) error {
	wd, err := syscall.InotifyAddWatch(w.watchfd, path, flags)
	if err != nil {
		return err
	}
	watch := int32(wd)
	w.watches[id] = watch
	w.ids[watch] = id
	return nil
}
func (w *inotify) Close() error {
	w.Lock()
	defer w.Unlock()
	if w.watchfd == -1 {
		return nil
	}
	if len(w.watches) == 0 {
		if err := w.add(NewId("/"), "/", deleteMask); err != nil {
			return err
		}
	}
	w.done <- true
	for id := range w.watches {
		w.remove(id)
	}
	return nil

}
func (w *inotify) remove(id Id) error {
	watch, ok := w.watches[id]
	if !ok {
		return fmt.Errorf("can't remove non-existent inotify watch for: %x", id)
	}
	success, errno := syscall.InotifyRmWatch(w.watchfd, uint32(watch))
	if success == -1 {
		return os.NewSyscallError("inotify_rm_watch", errno)
	}
	delete(w.watches, id)
	return nil
}
func (w *inotify) readEvents() {
	var buf [syscall.SizeofInotifyEvent * 4096]byte
	for {
		n, err := syscall.Read(w.watchfd, buf[:])
		var done bool
		select {
		case done = <-w.done:
		default:
		}
		if n == 0 || done {
			err = syscall.Close(w.watchfd)
			if err != nil {
				log.Println(os.NewSyscallError("close", err))
			}
			w.watchfd = -1
			return
		}
		if n < 0 {
			log.Println("inotify: read", err)
			continue
		}
		if n < syscall.SizeofInotifyEvent {
			log.Println("inotify: short read in readEvents()")
			continue
		}

		var offset uint32 = 0
		// We don't know how many events we just read into the buffer
		// While the offset points to at least one whole event...
		for offset <= uint32(n-syscall.SizeofInotifyEvent) {
			// Point "raw" to the event in the buffer
			raw := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
			// If the event happened to the watched directory or the watched file, the kernel
			// doesn't append the filename to the event, but we would like to always fill the
			// the "Name" field with a valid filename. We retrieve the path of the watch from
			// the "paths" map.
			w.Lock()
			id, ok := w.ids[raw.Wd]
			if !ok {
				log.Println("inotify: no resource found with id", id)
			}
			// Check if the the watch was removed
			if raw.Mask&syscall.IN_IGNORED != 0 {
				// remove stale watch
				delete(w.watches, id)
				delete(w.ids, raw.Wd)
			}
			w.Unlock()
			var name string
			if raw.Len > 0 {
				// Point "bytes" at the first byte of the filename
				bytes := (*[syscall.PathMax]byte)(unsafe.Pointer(&buf[offset+syscall.SizeofInotifyEvent]))
				// The filename is padded with NUL bytes. TrimRight() gets rid of those.
				name = strings.TrimRight(string(bytes[0:raw.Len]), "\000")
			}
			var op Op
			switch {
			case raw.Mask&createMask != 0 && name != "":
				op = Create
			case raw.Mask&modifyMask != 0 && name != "":
				op = Modify
			case raw.Mask&deleteMask != 0:
				op = Delete
			}
			if op != 0 {
				if err = w.ctrler.Control(op, id, name); err != nil {
					log.Println(err)
				}
			} // else log unexpected?
			// Move to the next event in the buffer
			offset += syscall.SizeofInotifyEvent + raw.Len
		}
	}
}
