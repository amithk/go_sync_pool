package syncpool

import "time"
import "sync/atomic"

type Channel struct {
	ch     chan interface{}
	size   int64
	hwmPct float64
	lwmPct float64
	closed int32
}

func NewChannel(size int64, hwmPct, lwmPct float64) *Channel {
	c := &Channel{
		ch:     make(chan interface{}, size),
		size:   size,
		hwmPct: hwmPct,
		lwmPct: lwmPct,
	}
	go c.Garbager()
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

func (c *Channel) Garbager() {
	for {
		if atomic.LoadInt32(&c.closed) == 1 {
			return
		}

		l := len(c.ch)
		if float64(l)/float64(c.size)*100 > c.hwmPct {
			// TODO: Need to implement garbaging in batches
			<-c.ch
		}
	}

	// TODO: Sleep less, maybe
	time.Sleep(1 * time.Millisecond)
}
