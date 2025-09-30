package storage


type KvStore interface {
	Put(k string, v string) error
	Get(k string) (string, error)
	Del(k string) error
	Close() error
}

func Engineerfactory(name string,dbpath string) KvStore {
	switch name {
	case "leveldb":
		levelDB,err := MakeLevelDBKvStore(dbpath)
		if err != nil {
			panic(err)
		}
		return levelDB
	default:
		return nil
	}

}
