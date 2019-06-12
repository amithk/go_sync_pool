package syncpool

import "errors"
import "math"
import "sync/atomic"
import "time"

import gometrics "github.com/rcrowley/go-metrics"

// Defaults
const (
	DEFAULT_CHAN_SIZE = 128
)

var (
	TEN_SECOND_EWMA = 1 - math.Exp(-5.0/10.0)
)

// Errors
var ErrorEmpty = errors.New("Pool is Empty")
var ErrorFull = errors.New("Pool is Full")
var ErrorClosed = errors.New("Pool is closed")

// Function to allocate new objects
type NewFunc func() interface{}

type ReturnCallback func() bool

// --------------------------------
// Pool
// --------------------------------
type Pool struct {
	New           NewFunc
	impl          PoolImpl
	enableCleaner bool
	cleanerStopCh chan bool
	inuse         int64
	inpool        int64
	ewma          gometrics.EWMA
	maxSize       int64

	// Max number of in-use objects in last interval
	// This is a better representation of expected value
	maxUsed int64
}

// -------------------------------
// Constructors start
// -------------------------------

func NewPoolWithImpl(New NewFunc, impl PoolImpl, enableCleaner bool, sz int64) *Pool {
	var stopCh chan bool

	if enableCleaner {
		stopCh = make(chan bool)
	}

	p := &Pool{
		New:           New,
		impl:          impl,
		enableCleaner: enableCleaner,
		cleanerStopCh: stopCh,
		ewma:          gometrics.NewEWMA(TEN_SECOND_EWMA),
		maxSize:       sz,
	}

	if enableCleaner {
		go p.Cleaner()
	}
	return p
}

func NewPool(fn NewFunc) *Pool {
	return NewSizedPoolWithChannel(fn, int64(DEFAULT_CHAN_SIZE))
}

func NewSizedPool(fn NewFunc, size int64) *Pool {
	return NewSizedPoolWithChannel(fn, size)
}

func NewPoolWithStack(fn NewFunc) *Pool {
	s := NewStack(-1)
	return NewPoolWithImpl(fn, s, true, -1)
}

func NewSizedPoolWithStack(fn NewFunc, size int64) *Pool {
	s := NewStack(size)
	return NewPoolWithImpl(fn, s, true, size)
}

func NewPoolWithQueue(fn NewFunc) *Pool {
	q := NewQueue(-1)
	return NewPoolWithImpl(fn, q, true, -1)
}

func NewSizedPoolWithQueue(fn NewFunc, size int64) *Pool {
	q := NewQueue(size)
	return NewPoolWithImpl(fn, q, true, size)
}

func NewSizedPoolWithChannel(fn NewFunc, size int64) *Pool {
	ch := NewChannel(size)
	return NewPoolWithImpl(fn, ch, true, size)
}

// -------------------------------
// Constructors end
// -------------------------------

func (p *Pool) Get() interface{} {
	val, err := p.impl.Get()
	if err == nil {
		atomic.AddInt64(&p.inuse, 1)
		atomic.AddInt64(&p.inpool, -1)
		p.updateMaxUsed()
		return val
	} else if err == ErrorEmpty {
		atomic.AddInt64(&p.inuse, 1)
		p.updateMaxUsed()
		return p.New()
	} else {
		// Unexpected error occurred
		panic(err)
	}
}

func (p *Pool) Put(val interface{}) {
	err := p.impl.Put(val)
	if err == nil {
		atomic.AddInt64(&p.inpool, 1)
		atomic.AddInt64(&p.inuse, -1)
	} else if err == ErrorFull {
		atomic.AddInt64(&p.inuse, -1)
	} else {
		// Unexpected error occurred
		panic(err)
	}
}

func (p *Pool) Close() {
	if p.enableCleaner {
		close(p.cleanerStopCh)
	}
	p.impl.Close()
}

func (p *Pool) Update() {
	maxUsed := atomic.LoadInt64(&p.maxUsed)
	p.ewma.Update(maxUsed)

	// reset maxUsed for the next interval
	atomic.AddInt64(&p.maxUsed, -maxUsed)
	p.ewma.Tick()
}

func (p *Pool) updateMaxUsed() {
	inuse := atomic.LoadInt64(&p.inuse)
	maxUsed := atomic.LoadInt64(&p.maxUsed)

	if maxUsed < inuse {
		atomic.AddInt64(&p.maxUsed, inuse-maxUsed)
	}
}

func (p *Pool) Clean() {
	exp := p.ewma.Rate()
	inpool := atomic.LoadInt64(&p.inpool)
	inuse := atomic.LoadInt64(&p.inuse)
	currCount := inpool + inuse

	if float64(currCount) > float64(exp)*1.1 {
		// Return some buffers from the pool
		// number of bufferes to be returned <= 10% of poolsize

		// TODO: If exp count remains zero for some time,
		// let the pool become completely empty
		var maxRetCount float64
		if p.maxSize < 0 {
			// Be lazy in cleaning
			maxRetCount = float64(1)
		} else {
			maxRetCount = float64(0.1) * float64(p.maxSize)
		}
		for i := 0; i < int(maxRetCount); i++ {
			removed := p.impl.RemoveOne()
			if !removed {
				return
			}
			atomic.AddInt64(&p.inpool, -1)
		}
	}
}

// For Cleaner to be effective, pool size has to be at least 10.
// With pool size less than 10, cleaner doesn't free the objects
// stored in the pool.
func (p *Pool) Cleaner() {
	i := 0
	for {
		select {
		case <-p.cleanerStopCh:
			return

		default:
		}

		// Update happens every 5 seconds
		// Note: this cannot change as go-metrics/ewma expects it like that
		if i%5 == 4 {
			p.Update()
		}

		// Clean the pool every 10 seconds.
		if i%10 == 9 {
			p.Clean()
		}

		i = (i + 1) % 10
		time.Sleep(1 * time.Second)
	}
}
