// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"runtime"
	"sync"
	"time"
)

// MountAll mounts all paths into workspace w.
// Paths are mounted in parallel if possible.
func MountAll(w *Ws, paths []string) []error {
	errs := make([]error, len(paths))
	if runtime.GOMAXPROCS(0) == 1 {
		for i, path := range paths {
			_, errs[i] = w.Mount(path)
		}
	} else {
		var wg sync.WaitGroup
		wg.Add(len(paths))
		mount := func(path string, err *error) {
			_, *err = w.Mount(path)
			wg.Done()
		}
		for i, path := range paths {
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

// Queue implements a locked resource queue.
type Queue struct {
	sync.Mutex
	queue []*Res
}

// Delete dequeues the resource.
func (q *Queue) Delete(r *Res) {
	q.Lock()
	defer q.Unlock()
	q.del(r)
}

// Add enqueues the resource.
// Resources already enqueued move to the end.
func (q *Queue) Add(r *Res) {
	q.Lock()
	defer q.Unlock()
	q.del(r)
	q.queue = append(q.queue, r)
}

// Work returns enqueued resources.
func (q *Queue) Work() []*Res {
	q.Lock()
	defer q.Unlock()
	res := make([]*Res, len(q.queue))
	copy(res, q.queue)
	q.queue = q.queue[:0]
	return res
}

func (q *Queue) del(r *Res) {
	for i, qr := range q.queue {
		if qr.Id == r.Id {
			q.queue = append(q.queue[:i], q.queue[i+1:]...)
			return
		}
	}
}

// Throttle implements a ticking resource queue.
// Tickers are created with Ticks, started, and sent to the Tickers channel.
type Throttle struct {
	sync.Mutex
	Queue
	Tickers chan *time.Ticker
	Ticks   time.Duration
	ticker  *time.Ticker
}

// Add enqueues the resource and starts a ticker if necessary.
func (q *Throttle) Add(r *Res) {
	q.Lock()
	defer q.Unlock()
	q.Queue.Add(r)
	if q.ticker == nil {
		q.ticker = time.NewTicker(q.Ticks)
		q.Tickers <- q.ticker
	}
}

// Work returns enqueued resources and stops the ticker if necessary.
func (q *Throttle) Work() []*Res {
	res := q.Queue.Work()
	q.Lock()
	defer q.Unlock()
	if len(q.queue) == 0 && q.ticker != nil {
		q.ticker.Stop()
		q.ticker = nil
	}
	return res
}
