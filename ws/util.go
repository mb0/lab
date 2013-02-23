package ws

import (
	"runtime"
	"sync"
	"time"
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

type Queue struct {
	sync.Mutex
	queue []*Res
	batch []*Res
}

func (q *Queue) del(r *Res) {
	for i, qr := range q.queue {
		if qr.Id == r.Id {
			q.queue = append(q.queue[:i], q.queue[i+1:]...)
			return
		}
	}
	for i, qr := range q.batch {
		if qr.Id == r.Id {
			q.batch = append(q.batch[:i], q.batch[i+1:]...)
			return
		}
	}
}
func (q *Queue) Delete(r *Res) {
	q.Lock()
	defer q.Unlock()
	q.del(r)
}
func (q *Queue) Add(r *Res) {
	q.Lock()
	defer q.Unlock()
	q.del(r)
	q.queue = append(q.queue, r)
}
func (q *Queue) Work() []*Res {
	q.Lock()
	defer q.Unlock()
	res := make([]*Res, len(q.batch))
	copy(res, q.batch)
	q.queue, q.batch = q.batch[:0], q.queue
	return res
}

type Throttle struct {
	sync.Mutex
	Queue
	Tickers chan *time.Ticker
	Ticks   time.Duration
	ticker  *time.Ticker
}

func (q *Throttle) Add(r *Res) {
	q.Lock()
	defer q.Unlock()
	q.Queue.Add(r)
	if q.ticker == nil {
		q.ticker = time.NewTicker(q.Ticks)
		q.Tickers <- q.ticker
	}
}
func (q *Throttle) Work() []*Res {
	q.Lock()
	defer q.Unlock()
	res := q.Queue.Work()
	if len(q.batch) == 0 && q.ticker != nil {
		q.ticker.Stop()
		q.ticker = nil
		q.Tickers <- nil
	}
	return res
}
