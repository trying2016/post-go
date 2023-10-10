package utils

import (
	"sync"
	"sync/atomic"
)

type QueueFun func(data interface{})

type Queue struct {
	queue      chan interface{}
	callBack   func(data interface{})
	exitSignal chan struct{}
	jobWaiter  *sync.WaitGroup
	count      *int64
}

func NewQueue(size int, fun func(data interface{})) *Queue {
	q := &Queue{}
	q.init(size, fun)
	return q
}

func (q *Queue) init(size int, fun func(data interface{})) {
	q.callBack = fun
	q.exitSignal = make(chan struct{})
	q.queue = make(chan interface{}, size)
	q.jobWaiter = &sync.WaitGroup{}
	q.jobWaiter.Add(1)
	var c int64
	q.count = &c
	go q.run()
}

func (q *Queue) Push(value interface{}) {
	atomic.AddInt64(q.count, 1)
	q.queue <- value
}
func (q *Queue) Count() int64 {
	return *q.count
}

func (q *Queue) Destroy() {
	q.exitSignal <- struct{}{}
	q.jobWaiter.Wait()
	close(q.queue)
}

func (q *Queue) run() {
	for {
		select {
		case data := <-q.queue:
			atomic.AddInt64(q.count, -1)

			q.callBack(data)
		case <-q.exitSignal:
			q.jobWaiter.Done()
			return
		}
	}
}
