package raftcore

import(
	"sync"
	"encoding/gob"
	"bytes"
	"encoding/binary"

	pb "neweraft/raftpb"
	"neweraft/storage"
)

type RaftLog struct{
	mu sync.RWMutex
	
	firstIndex int64
	lastIndex int64

	dbeng storage.KvStore
	entries []*pb.Entry
}

type RaftPersistentState struct{
	CurTerm int64
	VoteFor int64
	AppliedIdx int64
}

func MakeRaftLog(newdbeng storage.KvStore) *RaftLog{
	newRaftLog:=&RaftLog{
		firstIndex:0,
		lastIndex:1,
		dbeng:newdbeng,
		entries:[]*pb.Entry{},
	}
	newRaftLog.LoadPersistentLog()
	return newRaftLog
} 

func (rflog *RaftLog)LoadPersistentLog() error{
	rflog.mu.Lock()
	defer rflog.mu.Unlock()
	firstIdxByte,err:=rflog.dbeng.GetByte(append(RaftLogPrefix,FirstIdxKey...))
	if err!=nil{
		firstIdxByte=rflog.Int64toBytes(0)
	}
	firstIdx:= int64(binary.BigEndian.Uint64(firstIdxByte))
	lastIdxByte,err:=rflog.dbeng.GetByte(append(RaftLogPrefix,LastIdxKey...))
	if err!=nil{
		lastIdxByte=rflog.Int64toBytes(0)
	}
	lastIdx:= int64(binary.BigEndian.Uint64(lastIdxByte))
	newEntries:=[]*pb.Entry{}
	for i:=int64(0);i<=lastIdx-firstIdx;i++{
		IdxByte :=rflog.Int64toBytes(i+firstIdx)
		newEntryByte,errB :=rflog.dbeng.GetByte(append(RaftLogPrefix,IdxByte...))
		if errB!=nil {
			newEntryByte=[]byte{}
		}
		newEntry,err :=rflog.EntryDecode(newEntryByte)
		if err!=nil{
			newEntry  =&pb.Entry{}
		}
		newEntries=append(newEntries,newEntry)
	}
	rflog.firstIndex=firstIdx
	rflog.lastIndex=lastIdx
	rflog.entries=newEntries
	return nil
}

func (rflog *RaftLog)Int64toBytes(newInt64 int64) []byte{
	buf := make([]byte,8)
	binary.BigEndian.PutUint64(buf,uint64(newInt64))
	return buf
}

func (rflog *RaftLog)BytestoInt64(newBytes []byte) int64 {
	return int64(binary.BigEndian.Uint64(newBytes))
}

func (rflog *RaftLog)UpdatePersistentIndex(firstIdx int64,lastIdx int64) {
	bufF:= make([]byte,8)
	binary.BigEndian.PutUint64(bufF,uint64(firstIdx))
	rflog.dbeng.PutByte(append(RaftLogPrefix,FirstIdxKey...),bufF)
	bufL:= make([]byte,8)
	binary.BigEndian.PutUint64(bufL,uint64(lastIdx))
	rflog.dbeng.PutByte(append(RaftLogPrefix,LastIdxKey...),bufL)
}

func (rflog *RaftLog)EntryEncode(entry *pb.Entry) ([]byte,error){
	var buf bytes.Buffer
	enc :=gob.NewEncoder(&buf)
	err := enc.Encode(entry)
	if err!=nil {
		return []byte{},err
	}
	return buf.Bytes(),nil
}

func (rflog *RaftLog)EntryDecode(entryByte []byte) (*pb.Entry,error){
	buf :=bytes.NewBuffer(entryByte)
	dec := gob.NewDecoder(buf)
	entry := &pb.Entry{} 
	err := dec.Decode(entry)
	if err!=nil {
		return entry,err
	}
	return entry,nil
}

func (rflog *RaftLog)GetFirstIdx() int64{
	newFirstIdxByte,err:=rflog.dbeng.GetByte(append(RaftLogPrefix,FirstIdxKey...))
	if err !=nil{
		return 0
	}
	newFirstIdx := int64(binary.BigEndian.Uint64(newFirstIdxByte))
	return newFirstIdx
}

func (rflog *RaftLog)GetLastIdx() int64{
	newLastIdxByte,err:=rflog.dbeng.GetByte(append(RaftLogPrefix,LastIdxKey...))
	if err!= nil{
		return 0
	}
	newLastIdx := int64(binary.BigEndian.Uint64(newLastIdxByte))
	return newLastIdx
}

func (rflog *RaftLog)GetLastTerm() int64{
	lastIdxByte,LIBerr:=rflog.dbeng.GetByte(append(RaftLogPrefix,LastIdxKey...))
	if LIBerr!=nil{
		return 0
	}
	newEntryByte,NEBerr := rflog.dbeng.GetByte(lastIdxByte)
	if NEBerr !=nil{
		return 0
	}
	newEntry,_ :=rflog.EntryDecode(newEntryByte)
	return newEntry.CurTerm
}

func (rflog *RaftLog)GetEntries(fIdx int64,lIdx int64) ([]*pb.Entry){
	newEntries:=[]*pb.Entry{}
	for i:=int64(0);i<=lIdx-fIdx;i++{
		IdxByte:=rflog.Int64toBytes(fIdx+i)
		newEntryByte,err:=rflog.dbeng.GetByte(append(RaftLogPrefix,IdxByte...))
		var newEntry *pb.Entry
		if err!=nil{
			newEntry=&pb.Entry{}
		} else {
			newEntry,_=rflog.EntryDecode(newEntryByte)
		}
		newEntries=append(newEntries,newEntry)
	}
	return newEntries
}

func (rflog *RaftLog)GetEntry(idx int64) *pb.Entry{
	idxByte:=rflog.Int64toBytes(idx)
	newEntryByte,err:=rflog.dbeng.GetByte(append(RaftLogPrefix,idxByte...))
	var newEntry *pb.Entry
	if err!=nil{
		newEntry=&pb.Entry{}
	} else {
		newEntry,_=rflog.EntryDecode(newEntryByte)
	}
	return newEntry
}

func (rflog *RaftLog)AppendLogEntries(newEntries []*pb.Entry) error{
	for _,newEntry:= range newEntries{
		newEntryByte,_:=rflog.EntryEncode(newEntry)
		newIndexByte:=rflog.Int64toBytes(newEntry.Index)
		rflog.dbeng.PutByte(append(RaftLogPrefix,newIndexByte...),newEntryByte)
		rflog.dbeng.PutByte(append(RaftLogPrefix,LastIdxKey...),newIndexByte)

		rflog.entries= append(rflog.entries,newEntry)
		rflog.lastIndex = newEntry.Index
	}

	return nil
}