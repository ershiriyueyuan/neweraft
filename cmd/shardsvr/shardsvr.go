package main 

import(
	"os"
	"log"
	"strings"
	"strconv"
	"net"
    
	"google.golang.org/grpc"
	"neweraft/shardkvserver"
	pb "neweraft/raftpb"
)

func main(){
	if len(os.Args) < 3 {
		log.Println("输入格式:[id] [addr,addr,addr]")
		return
	}

	idStr:=os.Args[1]
	id,iderr:=strconv.Atoi(idStr)
	if iderr!=nil{
		log.Println("id atoi err")
	}

	addrs:=strings.Split(os.Args[2],",")
	peersAddrsMap:=make(map[int]string)
	for i,addr:=range addrs{
		peersAddrsMap[i]=addr
	}

	lis,liserr:=net.Listen("tcp",peersAddrsMap[id])
	if liserr!=nil{
		log.Fatalf("tcpcon: %v", liserr)
		return 
	}

	s:=grpc.NewServer()
	srdSvr:=shardkvserver.MakeShardServer(peersAddrsMap,int64(id))
	pb.RegisterMessageServiceServer(s,srdSvr)
	serverr:=s.Serve(lis)
	if serverr!=nil {
		log.Println("serve err")
	}
}