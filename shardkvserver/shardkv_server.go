package shardkvserver

import(
	"sync"
	"context"
	
	"neweraft/raftcore"
	pb "neweraft/raftpb"
)

type ShardServer struct{
	mu sync.RWMutex

	id int64

	raft *raftcore.Raft

	pb.UnimplementedMessageServiceServer
}

func MakeShardServer(peersAddrsMap map[int]string,idMe int64) *ShardServer{
	peers:=[]*raftcore.RaftClient{}
	for id,addr:=range peersAddrsMap {
		peer:=raftcore.MakeRaftClient(addr,int64(id))
		peers=append(peers,peer)
	}

	raft:=raftcore.MakeRaft(idMe,peers)

	shardServer:=&ShardServer{
		raft:raft,
	}

	return shardServer
}

func (shardsvr *ShardServer)RequestVote(ctx context.Context,req *pb.VoteRequest) (*pb.VoteResponse,error){
	res:=&pb.VoteResponse{}
	shardsvr.raft.HandleRequestVote(req,res)

	return res,nil
}

func (shardsvr *ShardServer)AppendEntry(ctx context.Context,req *pb.AppendEntryRequest) (*pb.AppendEntryResponse,error){
	res:=&pb.AppendEntryResponse{}
	shardsvr.raft.HandleAppendEntry(req,res)

	return res,nil
}