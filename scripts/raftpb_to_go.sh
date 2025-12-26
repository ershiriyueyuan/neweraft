go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc -I ../pbs \
       --go_out=../raftpb \
       --go-grpc_out=../raftpb \
       ../pbs/*.proto