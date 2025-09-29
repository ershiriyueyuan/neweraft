package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
)

type LevelDBKvStore struct {
	Path string
	db *leveldb.DB
}

func makeLevelDBKvStore() *LevelDBKvStore {
	return &LevelDBKvStore{
		Path: path,
		db: nil,
	}
}
