package raftcore

import(
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "neweraft/raftpb"
)

type RaftClient struct{
	id int64
	addr string
	conn *grpc.ClientConn
	MessageServiceClient pb.MessageServiceClient
}

func (raftcli *RaftClient) GetId() int64 {
	return raftcli.id
}

func (raftCli *RaftClient) GetMessageService() pb.MessageServiceClient{
	return raftCli.MessageServiceClient
}

func MakeRaftClient(addrMe string,idMe int64) *RaftClient{
	connMe,err:=grpc.NewClient(addrMe,grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err!=nil{
		log.Println("RaftClient conn err")
	}
	messageServiceClientMe:=pb.NewMessageServiceClient(connMe)
	return &RaftClient{
		id:idMe,
		addr:addrMe,
		conn:connMe,
		MessageServiceClient:messageServiceClientMe,
	}
}