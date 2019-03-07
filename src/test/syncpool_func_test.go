package test

import "testing"
import "syncpool"

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

func TestCleaner(t *testing.T) {

}
