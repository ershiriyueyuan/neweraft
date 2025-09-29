package storage

type KvStore interface {
	Put(k string, v string) error
	Get(k string) (string, error)
	Del(k string) error
}

func Engineerfactory(name string) KvStore {
	switch name {
	case "leveldb":
		return makeLevelDBKvStore()
	default:
		return nil
	}
}
