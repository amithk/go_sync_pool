package syncpool

type Channel struct {
	ch chan interface{}
}

func NewChannel(size int64) *Channel {
	return &Channel{
		ch: make(chan interface{}, size),
	}
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
}
