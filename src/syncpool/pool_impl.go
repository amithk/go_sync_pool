package syncpool

type PoolImpl interface {
	Get() (interface{}, error)

	Put(interface{}) error

	Close()

	// Remove and discard 'count' number of elements
	Remove(count int64)
}
