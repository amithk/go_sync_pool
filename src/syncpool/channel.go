package syncpool

import "sync/atomic"

type Channel struct {
	ch     chan interface{}
	size   int64
	closed int32
}

func NewChannel(size int64) *Channel {
	c := &Channel{
		ch:   make(chan interface{}, size),
		size: size,
	}
	return c
}

func (c *Channel) Get() (interface{}, error) {
	var i interface{}
	var ok bool
	select {
	case i, ok = <-c.ch:
		if !ok {
			return nil, ErrorClosed
		}
		break

	default:
		return nil, ErrorEmpty
	}
	return i, nil
}

func (c *Channel) Put(i interface{}) error {
	select {
	case c.ch <- i:
		break

	default:
		return ErrorFull
	}
	return nil
}

func (c *Channel) Close() {
	// TODO: Need to avoid closing of already closed channel
	// close(c.ch)
	atomic.StoreInt32(&c.closed, 1)
}

func (c *Channel) Remove(count int64) {
	// TODO: Implementation needed
}
