package syncpool

type PoolImpl interface {
	Get() (interface{}, error)

	Put(interface{}) error

	Close()

	// Remove and discard one element
	// Returns true if removed, else false
	RemoveOne() bool
}
