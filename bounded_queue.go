package queue

import (
	"sync"
)

type BoundedQueue struct {
	cap      int
	s        []interface{}
	mutex    *sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond
	front    int
	rear     int
}

func NewBoundedQueue(cap int) *BoundedQueue {
	mutex := &sync.Mutex{}
	notEmpty := sync.NewCond(mutex)
	notFull := sync.NewCond(mutex)
	s := make([]interface{}, cap)
	return &BoundedQueue{cap: cap, s: s,
		mutex: mutex, notEmpty: notEmpty, notFull: notFull}
}

func (o *BoundedQueue) Push(v interface{}) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	for o.full() {
		o.notFull.Wait()
	}
	o.s[o.rear] = v
	o.rear = (o.rear + 1) % o.cap
	o.notEmpty.Broadcast()
}

// LRU
func (o *BoundedQueue) ForcePush(v interface{}) interface{} {
	var poped interface{}
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.full() {
		poped = o.s[o.front]
		o.front = (o.front + 1) % o.cap
	}
	o.s[o.rear] = v
	o.rear = (o.rear + 1) % o.cap
	o.notEmpty.Broadcast()

	return poped
}

func (o *BoundedQueue) Pop() interface{} {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	for o.empty() {
		o.notEmpty.Wait()
	}
	v := o.s[o.front]
	o.front = (o.front + 1) % o.cap
	o.notFull.Broadcast()

	return v
}

func (o *BoundedQueue) Full() bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.full()
}

func (o *BoundedQueue) Empty() bool {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.empty()
}

func (o *BoundedQueue) Cap() int {
	return o.cap
}

func (o *BoundedQueue) Len() int {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	return o.len()
}

func (o *BoundedQueue) Walk(f func(v interface{})) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	for i := 0; i < o.len(); i++ {
		idx := (o.front + i) % o.cap
		f(o.s[idx])
	}
}

func (o *BoundedQueue) full() bool {
	return (o.rear+1)%o.cap == o.front
}

func (o *BoundedQueue) empty() bool {
	return o.front == o.rear
}

func (o *BoundedQueue) len() int {
	rear := o.rear
	if rear < o.front {
		rear += o.cap
	}
	return rear - o.front
}
