package storage

import (
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type LevelDBKvStore struct {
	Path string
	db *leveldb.DB
}

func MakeLevelDBKvStore(dbpath string) (*LevelDBKvStore,error) {
	db,err:=leveldb.OpenFile(dbpath,nil)
	if err != nil {
    return nil, err  // 必须处理错误
	}	
	return &LevelDBKvStore{
		Path: dbpath,
		db: db,
	},err
}

func (l *LevelDBKvStore) Close() error {
	if l.db != nil {
		return l.db.Close()
	}
	return nil
}

func (l *LevelDBKvStore) Put(k string,v string) error {
	if l.db == nil {
		return errors.New("database not opened")
	}
	return l.db.Put([]byte(k), []byte(v),nil)
}

func (l *LevelDBKvStore) Get(k string) (string,error) {
	if l.db ==nil {
		return "",errors.New("database not opened")
	}
	data,err := l.db.Get([]byte(k),nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return "",errors.New("Key not found")
		}
		return "",err
	}
	return string(data),nil
}

func (l *LevelDBKvStore) Del(k string) error {
	if l.db ==nil {
		return errors.New("database not opened")
	}
	return l.db.Delete([]byte(k),nil)
}
