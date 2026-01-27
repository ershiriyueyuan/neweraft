package storage

import (
	"errors"
	"strings"
	"encoding/binary"
	
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
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

func (l *LevelDBKvStore) PutByte(k []byte,v []byte) error{
	return l.db.Put(k,v,nil)
}

func (l *LevelDBKvStore) GetByte(k []byte) ([]byte, error){
	return l.db.Get(k,nil)
}

func (l *LevelDBKvStore) DelByte(k []byte) error{
	return l.db.Delete(k,nil)
}

func (l *LevelDBKvStore) SeekPrefixFirst(prefix string) ([]byte,[]byte,error){
	iter:=l.db.NewIterator(util.BytesPrefix([]byte(prefix)),nil)
	defer iter.Release()
	if iter.Next(){
		return iter.Key(),iter.Value(),nil
	}
	return []byte{}, []byte{}, errors.New("seek not find key")
}

func (l *LevelDBKvStore) DumpPrefix(prefix string,trimPrefix bool) (map[string]string,error){
	iter:=l.db.NewIterator(util.BytesPrefix([]byte(prefix)),nil)
	defer iter.Release()
	kvMap:=map[string]string{}
	for iter.Next(){
		k:=iter.Key()
		if trimPrefix{
			k=[]byte(strings.TrimPrefix(string(k),prefix))
		}
		v:=iter.Value()
		kvMap[string(k)]=string(v)
	}
	return kvMap,nil
}


func (l *LevelDBKvStore) DelPrefix(prefix string) error{
	iter:=l.db.NewIterator(util.BytesPrefix([]byte(prefix)),nil)
	defer iter.Release()
	for iter.Next(){
		delerr:=l.db.Delete(iter.Key(),nil)
		if delerr!=nil{
			return delerr
		}
	}
	return nil
}

func (l *LevelDBKvStore) SeekPrefixLast(prefix []byte) ([]byte,[]byte,error){
	iter:=l.db.NewIterator(util.BytesPrefix(prefix),nil)
	defer iter.Release()
	if iter.Last() {
		return iter.Key(),iter.Value(),nil
	}
	return []byte{}, []byte{}, errors.New("seek not find key")
}

func (l *LevelDBKvStore) SeekPrefixIdmax(prefix []byte) (int64,error){
	iter:=l.db.NewIterator(util.BytesPrefix(prefix),nil)
	defer iter.Release()
	var maxKeyId int64
	maxKeyId =0
	for iter.Next(){
		if iter.Error() !=nil {
			return maxKeyId,iter.Error()
		}
		kBytes := iter.Key()
		kId:=int64(binary.LittleEndian.Uint64(kBytes[len(prefix):]))
		if kId>maxKeyId {
			maxKeyId=kId
		}
	}
	return maxKeyId,nil
}