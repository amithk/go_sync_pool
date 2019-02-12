package syncpool

import "errors"

// Defaults
var (
	DEFAULT_CHAN_SIZE = 128
)

// Errors
var ErrorEmpty = errors.New("Empty")
var ErrorFull = errors.New("Full")
var ErrorClosed = errors.New("Pool is closed")

// Function to allocate new objects
type NewFunc func() interface{}

type Pool struct {
	New  NewFunc
	impl PoolImpl
}

// -------------------------------
// Constructors start
// -------------------------------
func NewPool(fn NewFunc) *Pool {
	return NewSizedPoolWithChannel(fn, int64(DEFAULT_CHAN_SIZE))
}

func NewSizedPool(fn NewFunc, size int64) *Pool {
	return NewSizedPoolWithChannel(fn, size)
}

func NewPoolWithStack(fn NewFunc) *Pool {
	p := &Pool{
		New:  fn,
		impl: NewStack(-1),
	}
	return p
}

func NewSizedPoolWithStack(fn NewFunc, size int64) *Pool {
	p := &Pool{
		New:  fn,
		impl: NewStack(size),
	}
	return p
}

func NewPoolWithQueue(fn NewFunc) *Pool {
	p := &Pool{
		New:  fn,
		impl: NewQueue(-1),
	}
	return p
}

func NewSizedPoolWithQueue(fn NewFunc, size int64) *Pool {
	p := &Pool{
		New:  fn,
		impl: NewQueue(size),
	}
	return p
}

func NewSizedPoolWithChannel(fn NewFunc, size int64) *Pool {
	p := &Pool{
		New:  fn,
		impl: NewChannel(size),
	}
	return p
}

// -------------------------------
// Constructors end
// -------------------------------

func (p *Pool) Get() interface{} {
	val, err := p.impl.Get()
	if err == nil {
		return val
	} else if err == ErrorEmpty {
		return p.New()
	} else {
		panic(err)
	}
}

func (p *Pool) Put(val interface{}) {
	_ = p.impl.Put(val)
}
