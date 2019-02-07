package syncpool

import "sync/atomic"
import "unsafe"

type Stack struct {
	top   unsafe.Pointer
	size  int64
	count int64
}

func NewStack(size int64) *Stack {
	return &Stack{
		top:  nil,
		size: size,
	}
}

func (s *Stack) Put(val interface{}) error {

	size := atomic.LoadInt64(&s.size)
	if size >= 0 {
		sz := atomic.AddInt64(&s.count, 1)
		if sz > size {
			_ = atomic.AddInt64(&s.count, -1)
			return ErrorFull
		}
	}

	var swapped bool
	n := NewNode(val)
	n.next = nil
	un := unsafe.Pointer(n)

	for {
		t := atomic.LoadPointer(&s.top)
		if t == nil {
			n.prev = nil
			swapped = atomic.CompareAndSwapPointer(&s.top, nil, un)
			if !swapped {
				continue
			}
			return nil
		} else {
			n.prev = t
			swapped = atomic.CompareAndSwapPointer(&s.top, t, un)
			if !swapped {
				continue
			}
			return nil
		}
	}
	return nil
}

func (s *Stack) Get() (interface{}, error) {
	var swapped bool
	for {
		t := atomic.LoadPointer(&s.top)
		if t == nil {
			return nil, ErrorEmpty
		}

		nt := (*Node)(t)
		p := nt.prev
		val := nt.val
		if p == nil {
			swapped = atomic.CompareAndSwapPointer(&s.top, t, nil)
			if !swapped {
				continue
			}
			_ = atomic.AddInt64(&s.count, -1)
			return val, nil
		} else {
			swapped = atomic.CompareAndSwapPointer(&s.top, t, p)
			if !swapped {
				continue
			}

			np := (*Node)(p)
			swapped = atomic.CompareAndSwapPointer(&np.next, t, nil)
			if !swapped {
				continue
			}
			_ = atomic.AddInt64(&s.count, -1)
			return val, nil
		}
	}
	return nil, nil
}
