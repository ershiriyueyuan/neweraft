package storage


type KvStore interface {
	Put(k string, v string) error
	Get(k string) (string, error)
	Del(k string) error
	
	PutByte(k []byte,v []byte) error
	GetByte(k []byte) ([]byte, error)
	DelByte(k []byte) error

	SeekPrefixFirst(prefix string) ([]byte,[]byte,error)
	DumpPrefix(prefix string,trimPrefix bool) (map[string]string,error)
	DelPrefix(prefix string) error

	SeekPrefixLast(prefix []byte) ([]byte,[]byte,error)
	SeekPrefixIdmax(prefix []byte) (int64,error)
	
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
