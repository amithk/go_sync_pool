package syncpool

type PoolImpl interface {
	Get() (interface{}, error)

	Put(interface{}) error

	Close()
}
