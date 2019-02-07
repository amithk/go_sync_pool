package test

import "testing"
import "syncpool"
import "sync"
import "time"
import "fmt"

func doGetGoSyncPool(gp *sync.Pool) *[]byte {
	return gp.Get().(*[]byte)
}

func doPutGoSyncPool(gp *sync.Pool, buf *[]byte) {
	gp.Put(buf)
}

func doGetPutGoSyncPool(gp *sync.Pool) {
	for j := 0; j < 100000; j++ {
		buf := doGetGoSyncPool(gp)
		time.Sleep(1 * time.Nanosecond)
		doPutGoSyncPool(gp, buf)
	}
}

func doGetPutGoSyncPoolRoutine(gp *sync.Pool, donech chan bool) {
	doGetPutGoSyncPool(gp)
	donech <- true
}

func doGet(p *syncpool.Pool) *[]byte {
	b := p.Get()
	return b.(*[]byte)
}

func doPut(p *syncpool.Pool, buf *[]byte) {
	p.Put(buf)
}

func doGetPut(p *syncpool.Pool) {
	for i := 0; i < 100000; i++ {
		buf := doGet(p)
		time.Sleep(1 * time.Nanosecond)
		doPut(p, buf)
	}
}

func doGetPutRoutine(p *syncpool.Pool, donech chan bool) {
	doGetPut(p)
	donech <- true
}

func SpawnAndWait(p *syncpool.Pool, n int) {
	chans := make([]chan bool, n)
	for i := 0; i < n; i++ {
		ch := make(chan bool)
		chans[i] = ch
		go doGetPutRoutine(p, ch)
	}

	for j := 0; j < n; j++ {
		<-chans[j]
	}
}

func SpawnAndWaitGoSyncPool(gp *sync.Pool, n int) {
	chans := make([]chan bool, n)
	for i := 0; i < n; i++ {
		ch := make(chan bool)
		chans[i] = ch
		go doGetPutGoSyncPoolRoutine(gp, ch)
	}

	for j := 0; j < n; j++ {
		<-chans[j]
	}
}

func TestBasicGetPut(t *testing.T) {
	size := 16 * 1024
	newfn := func() interface{} {
		buf := make([]byte, 0, size)
		return &buf
	}

	p := syncpool.NewPool(newfn)

	// Get 3 buffers
	buf1 := p.Get()
	buf2 := p.Get()
	buf3 := p.Get()

	// Return 2 of them
	p.Put(buf1)
	p.Put(buf3)

	// Get 2 more buffers
	buf4 := p.Get()
	buf5 := p.Get()

	// Return all of the buffers
	p.Put(buf2)
	p.Put(buf5)
	p.Put(buf4)
}

func BenchmarkSingleThread(b *testing.B) {
	size := 16 * 1024
	newfn := func() interface{} {
		buf := make([]byte, 0, size)
		return &buf
	}

	p := syncpool.NewPool(newfn)
	gp := &sync.Pool{
		New: newfn,
	}

	st1 := time.Now()
	doGetPut(p)
	et1 := time.Now()
	fmt.Println("Time taken by New Sync Pool implementation =", et1.Sub(st1))

	st2 := time.Now()
	doGetPutGoSyncPool(gp)
	et2 := time.Now()
	fmt.Println("Time taken by Golang Sync Pool implementation =", et2.Sub(st2))
}

func Benchmark64Threads(b *testing.B) {
	size := 16 * 1024
	newfn := func() interface{} {
		buf := make([]byte, 0, size)
		return &buf
	}

	p := syncpool.NewPool(newfn)
	gp := &sync.Pool{
		New: newfn,
	}

	st1 := time.Now()
	SpawnAndWait(p, 64)
	et1 := time.Now()
	fmt.Println("Time taken by New Sync Pool implementation =", et1.Sub(st1))

	st2 := time.Now()
	SpawnAndWaitGoSyncPool(gp, 64)
	et2 := time.Now()
	fmt.Println("Time taken by Golang Sync Pool implementation =", et2.Sub(st2))
}
