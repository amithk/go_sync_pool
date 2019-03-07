package syncpool

import "sync/atomic"
import "unsafe"

type Queue struct {
	front unsafe.Pointer
	rear  unsafe.Pointer
	size  int64
	count int64
}

func NewQueue(size int64) *Queue {
	return &Queue{
		front: nil,
		rear:  nil,
		size:  size,
	}
}

func (q *Queue) Put(val interface{}) error {

	size := atomic.LoadInt64(&q.size)
	if size >= 0 {
		sz := atomic.AddInt64(&q.count, 1)
		if sz > size {
			_ = atomic.AddInt64(&q.count, -1)
			return ErrorFull
		}
	}

	var swapped bool
	n := NewNode(val)
	n.prev = nil
	un := unsafe.Pointer(n)

	for {
		r := atomic.LoadPointer(&q.rear)
		atomic.StorePointer(&n.next, r)

		swapped = atomic.CompareAndSwapPointer(&q.rear, r, un)
		if !swapped {
			continue
		}

		f := atomic.LoadPointer(&q.front)
		if f == nil {
			swapped = atomic.CompareAndSwapPointer(&q.front, nil, un)
			if !swapped {
				continue
			}
		} else {
			if r == nil {
				continue
			}

			nr := (*Node)(r)
			swapped = atomic.CompareAndSwapPointer(&nr.prev, nil, un)
			if !swapped {
				continue
			}
		}

		break
	}
	return nil
}

func (q *Queue) Get() (interface{}, error) {
	var swapped bool
	for {
		f := atomic.LoadPointer(&q.front)
		if f == nil {
			return nil, ErrorEmpty
		}

		nf := (*Node)(f)
		p := atomic.LoadPointer(&nf.prev)
		r := atomic.LoadPointer(&q.rear)

		val := nf.val
		if p == nil {
			swapped = atomic.CompareAndSwapPointer(&q.front, f, nil)
			if !swapped {
				continue
			}

			swapped = atomic.CompareAndSwapPointer(&q.rear, r, nil)
			if !swapped {
				continue
			}
			_ = atomic.AddInt64(&q.count, -1)
			return val, nil
		} else {
			np := (*Node)(p)
			swapped = atomic.CompareAndSwapPointer(&np.next, f, nil)
			if !swapped {
				continue
			}
			swapped = atomic.CompareAndSwapPointer(&q.front, f, p)
			if !swapped {
				continue
			}
			_ = atomic.AddInt64(&q.count, -1)
			return val, nil
		}
	}
}

func (q *Queue) Empty() bool {
	f := atomic.LoadPointer(&q.front)
	r := atomic.LoadPointer(&q.rear)

	if r == nil && f == nil {
		return true
	} else {
		return false
	}
}

func (q *Queue) Close() {
	// TODO: Drain the queue
}

func (q *Queue) RemoveOne() bool {
	var err error
	_, err = q.Get()
	if err == nil {
		return true
	} else if err == ErrorEmpty {
		return false
	} else {
		panic(err)
	}
}
