package raftcore

import(
	"log"
	"sync"
	"time"
	"context"
	"math/rand"
	"bytes"
	"encoding/gob"

	pb "neweraft/raftpb"
	"neweraft/storage"
)

type RaftRole int

const(
	RaftFollower RaftRole=iota
	RaftCandidate
	RaftLeader
)

type Raft struct {
	mu sync.RWMutex

	rflog *RaftLog
	logEng storage.KvStore
	nextIndexs []int64
	matchIndexs []int64
	commitIndex int64
	appliedIndex int64

	id int64
	leaderId int64
	deadIf bool
	role RaftRole
	curTerm int64
	peers []*RaftClient

	countVote int64
	voteFor int64
	electionTimer *time.Timer
	electionTime time.Duration

	heartTimer *time.Timer
	heartTime time.Duration
}

func MakeRaft(id int64,peers []*RaftClient,logeng storage.KvStore) *Raft{
	electionTime:=time.Duration(500 + rand.Intn(150)) * time.Millisecond
	heartTime:=100*time.Millisecond
	lenSize:=int64(len(peers))
	raft:=&Raft{
		id:id,
		role:RaftFollower,
		countVote:0,
		voteFor:-1,
		heartTimer:time.NewTimer(heartTime),
		electionTimer:time.NewTimer(electionTime),
		peers:peers,
		deadIf:false,
		curTerm:0,
		electionTime: electionTime,
		heartTime: heartTime,
		logEng:logeng,
		rflog:MakeRaftLog(logeng),
		matchIndexs:make([]int64, lenSize),
		nextIndexs:make([]int64, lenSize),
		commitIndex:0,
		leaderId:id,
	}
	newRaftPersistentState:=raft.GetPersistState()
	raft.curTerm=newRaftPersistentState.CurTerm
	raft.voteFor=newRaftPersistentState.VoteFor
	raft.appliedIndex=newRaftPersistentState.AppliedIdx
	
	raft.heartTimer.Stop()
	raft.electionTimer.Reset(raft.electionTime)
	go raft.Tick() 

	return raft
}

func(raft *Raft) Tick(){
	for !raft.isKill() {
		select{
		case <-raft.electionTimer.C:
			if(raft.role==RaftLeader){
				break
			}
			raft.switchRole(RaftCandidate)
			raft.electionTimer.Reset(raft.electionTime)
		case <-raft.heartTimer.C:
			raft.broadcastHeart()
			raft.heartTimer.Reset(raft.heartTime)
		}
	}
}

func (raft *Raft)switchRole(newRole RaftRole) {
	raft.mu.Lock()
	if raft.role==newRole{
		raft.mu.Unlock()
		return
	}
	log.Printf("Node %d state change: %d -> %d (Term %d)", raft.id, raft.role, newRole, raft.curTerm)
	raft.role=newRole
	raft.mu.Unlock()

	switch newRole{
	case RaftCandidate:
		raft.election()
	case RaftFollower:
		raft.heartTimer.Stop()	
	case RaftLeader:
		raft.electionTimer.Stop()	
		raft.heartTimer.Reset(raft.heartTime)
	}

	raft.MakePersistState()
	return
} 

func(raft *Raft) isKill() bool{
	return raft.deadIf
}


func (raft *Raft)broadcastHeart(){
	for _, peer:=range raft.peers{
		if(peer.id==raft.id){
			continue
		}

		go raft.replicateOneround(peer)
	}
}

func (raft *Raft)replicateOneround(peer *RaftClient) {
	if raft.role!=RaftLeader{
		return
	}



	appendEntryRequest:=&pb.AppendEntryRequest{
		CurTerm:raft.curTerm,
		LeaderId:raft.leaderId,
		PreLogIndex:raft.nextIndexs[peer.id]-1,
		PreLogTerm:raft.rflog.GetEntry(raft.nextIndexs[peer.id]-1).CurTerm,
		Entries:raft.rflog.GetEntries(raft.nextIndexs[peer.id],raft.rflog.GetLastIdx()),
	}
	ctx,cancel:=context.WithTimeout(context.Background(),200 * time.Millisecond)
	defer cancel()
	appendEntryResponse,err:=peer.MessageServiceClient.AppendEntry(ctx,appendEntryRequest)
	if err!=nil {
		peer=MakeRaftClient(peer.addr,peer.id)
		raft.peers[peer.id]=peer
		appendEntryResponse,_=peer.MessageServiceClient.AppendEntry(ctx,appendEntryRequest)
		// log.Printf("AppendEntryResponse %d error: %v",peer.id, err)
	} else {
		if appendEntryResponse.Success{
			raft.matchIndexs[peer.id]=appendEntryRequest.PreLogIndex
			raft.nextIndexs[peer.id]=raft.matchIndexs[peer.id]+1
			
			//commit通知

		} else {
			if appendEntryResponse.Term<raft.curTerm && appendEntryResponse.ConflictIndex<raft.rflog.GetFirstIdx() {
				raft.nextIndexs[peer.id]=appendEntryResponse.ConflictIndex
				raft.replicateOneround(peer)
			} else if appendEntryResponse.Term>raft.curTerm{
				raft.switchRole(RaftFollower)
			}
		}
	}	
}

func (raft *Raft)HandleRequestVote(req *pb.VoteRequest,res *pb.VoteResponse){
	if(raft.curTerm<req.CurTerm){
		raft.switchRole(RaftFollower)
		res.VoteGranted=true
	} else {
		res.VoteGranted=false   
	}

	raft.electionTimer.Reset(raft.electionTime)
}

func (raft *Raft)HandleAppendEntry(req *pb.AppendEntryRequest,res *pb.AppendEntryResponse){
	res.Term=raft.curTerm

	if(req.CurTerm<raft.curTerm){
		res.Success =false
		return
	}

	if req.CurTerm > raft.curTerm {
		raft.curTerm = req.CurTerm
		raft.voteFor = -1
		raft.switchRole(RaftFollower)
	}

	raft.electionTimer.Reset(raft.electionTime)

	if req.PreLogIndex==raft.rflog.GetLastIdx() && req.PreLogTerm==raft.rflog.GetLastTerm(){
		raft.rflog.AppendLogEntries(req.Entries)
		res.Success=true
		res.Term=req.CurTerm
		raft.curTerm=req.CurTerm
	} else {
		res.Success=false
		res.ConflictIndex=raft.rflog.GetLastIdx()
		res.ConflictTerm=raft.rflog.GetLastTerm()
	}
}

func(raft *Raft) election(){
	raft.mu.Lock()
	
	raft.curTerm++
	raft.voteFor = raft.id
	raft.countVote=1

	raft.mu.Unlock()


	for _,peer:=range raft.peers{
		if(peer.id==raft.id){
				continue
		}

		go func(p *RaftClient){
			voteRequest:=&pb.VoteRequest{
				CurTerm:raft.curTerm,
				SefId:raft.id,
			}

			ctx,cancel :=context.WithTimeout(context.Background(),200 * time.Millisecond)
			defer cancel()
			voteResponse,err:=p.MessageServiceClient.RequestVote(ctx,voteRequest)
			if err!=nil {
				log.Printf("voteResponse %d error: %v",p.id, err)
				return
			}

			if voteResponse.VoteGranted {
				raft.mu.Lock()
				raft.countVote++
				raft.mu.Unlock()
			}

			if raft.countVote>int64((len(raft.peers))/2) {
				raft.switchRole(RaftLeader)
			}
		}(peer)
	}
}

func (raft *Raft) MakePersistState() error{
	newPesistState:=&RaftPersistentState{
		CurTerm:raft.curTerm,
		VoteFor:raft.voteFor,
		AppliedIdx:raft.appliedIndex,
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(newPesistState)
	if err != nil{
		return err
	}
	raft.logEng.PutByte(RaftStateKey,buf.Bytes())
	return nil
}

func (raft *Raft) GetPersistState() *RaftPersistentState{
	RaftPersistentStateByte,_ := raft.logEng.GetByte(RaftStateKey)
	buf := bytes.NewBuffer(RaftPersistentStateByte)
	dec := gob.NewDecoder(buf)
	newRaftPersistentState:=&RaftPersistentState{}
	dec.Decode(newRaftPersistentState)
	return newRaftPersistentState
}