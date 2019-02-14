package syncpool

import "errors"
import "time"
import gometrics "rcrowley/go-metrics"

// Defaults
var (
	DEFAULT_CHAN_SIZE = 128
)

// Errors
var ErrorEmpty = errors.New("Pool is Empty")
var ErrorFull = errors.New("Pool is Full")
var ErrorClosed = errors.New("Pool is closed")

// Function to allocate new objects
type NewFunc func() interface{}

// --------------------------------
// GarbagerInfo
// --------------------------------

type GarbagerInfo struct {
	active int64
	free   int64
	ewma   gometrics.EWMA
}

func NewGarbagerInfo(ewma gometrics.EWMA) *GarbagerInfo {
	return &GarbagerInfo{ewma: ewma}
}

// --------------------------------
// Pool
// --------------------------------
type Pool struct {
	New            NewFunc
	impl           PoolImpl
	enableGarbager bool
	garbagerStopCh chan bool
	ginfo          *GarbagerInfo
}

func NewPoolWithImpl(New NewFunc, impl PoolImpl, enableGarbager bool) *Pool {
	var stopCh chan bool
	var ginfo *GarbagerInfo

	if enableGarbager {
		stopCh = make(chan bool)
		ginfo = NewGarbagerInfo(gometrics.NewEWMA1())
	}

	p := &Pool{
		New:            New,
		impl:           impl,
		enableGarbager: enableGarbager,
		garbagerStopCh: stopCh,
		ginfo:          ginfo,
	}

	if enableGarbager {
		go p.Garbager()
	}
	return p
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
	s := NewStack(-1)
	return NewPoolWithImpl(fn, s, true)
}

func NewSizedPoolWithStack(fn NewFunc, size int64) *Pool {
	s := NewStack(size)
	return NewPoolWithImpl(fn, s, true)
}

func NewPoolWithQueue(fn NewFunc) *Pool {
	q := NewQueue(-1)
	return NewPoolWithImpl(fn, q, true)
}

func NewSizedPoolWithQueue(fn NewFunc, size int64) *Pool {
	q := NewQueue(size)
	return NewPoolWithImpl(fn, q, true)
}

func NewSizedPoolWithChannel(fn NewFunc, size int64) *Pool {
	ch := NewChannel(size)
	return NewPoolWithImpl(fn, ch, true)
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

func (p *Pool) Close() {
	close(p.garbagerStopCh)
	p.impl.Close()
}

func (p *Pool) Garbager() {
	for {
		select {
		case <-p.garbagerStopCh:
			// TODO: Add log message
			return

		default:
			// TODO: Update the usage
			// TODO: If required, remove 10%
		}

		// 10 Second Interval
		time.Sleep(10 * time.Second)
	}
}
