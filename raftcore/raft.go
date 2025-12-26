package raftcore

import(
	"log"
	"sync"
	"time"
	"context"

	pb "neweraft/raftpb"
)

type RaftRole int

const(
	RaftFollower RaftRole=iota
	RaftCandidate
	RaftLeader
)

type Raft struct {
	mu sync.RWMutex

	id int64
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

func MakeRaft(addr string,id int64,peers []*RaftClient) *Raft{
	electionTime:=300*time.Millisecond
	heartTime:=50*time.Millisecond
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
	}

	go raft.Tick() 

	return raft
}

func(raft *Raft) Tick(){
	for !raft.isKill() {
		select{
		case <-raft.electionTimer.C:
			raft.switchRole(RaftCandidate)
			raft.election()
			raft.electionTimer.Reset(raft.electionTime)
		case <-raft.heartTimer.C:
			raft.broadcastHeart()
			raft.heartTimer.Reset(raft.heartTime)
		}
	}
}

func (raft *Raft)switchRole(newRole RaftRole) {

	if raft.role==newRole{
		return
	}
	raft.mu.Lock()
	raft.role=newRole
	raft.mu.Unlock()

	switch raft.role{
	case RaftCandidate:
		raft.election()
	case RaftFollower:
		raft.heartTimer.Stop()	
	case RaftLeader:
		raft.heartTimer.Reset(raft.heartTime)
	}

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
	appendEntryRequest:=&pb.AppendEntryRequest{
		CurTerm:raft.curTerm,
	}


	ctx,cancel:=context.WithTimeout(context.Background(),100 * time.Millisecond)
	defer cancel()
	appendEntryResponse,err:=peer.MessageServiceClient.AppendEntry(ctx,appendEntryRequest)
	if err!=nil {
		log.Printf("AppendEntryResponse %d error",peer.id)
	}

	_ = appendEntryResponse 
}

func (raft *Raft)HandleRequestVote(req *pb.VoteRequest,res *pb.VoteResponse){
	if(raft.curTerm<=req.CurTerm){
		res.VoteGranted=true
	}

	raft.electionTimer.Reset(raft.electionTime)
}

func (raft *Raft)HandleAppendEntry(){
	raft.electionTimer.Reset(raft.electionTime)
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

			ctx,cancel :=context.WithTimeout(context.Background(),100 * time.Millisecond)
			defer cancel()
			voteResponse,err:=p.MessageServiceClient.RequestVote(ctx,voteRequest)
			if err!=nil {
			log.Printf("voteResponse %d error",peer.id)
		}

			if voteResponse.VoteGranted {
				raft.countVote++
			}

			if raft.countVote>int64((len(raft.peers))/2) {
				raft.switchRole(RaftLeader)
			}
		}(peer)
	}
}