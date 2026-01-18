#!/bin/bash

cd ..

# 编译
go build -o ./out/shardsvr ./cmd/shardsvr/shardsvr.go

sleep 2

# 杀掉旧进程
fuser -k 8088/tcp 8089/tcp 8090/tcp 8091/tcp 8092/tcp

# 启动?个节点
echo "Starting servers..."
./out/shardsvr 0 :8088,:8089,:8090,:8091,:8092 &
PID0=$!
./out/shardsvr 1 :8088,:8089,:8090,:8091,:8092 &
PID1=$!
./out/shardsvr 2 :8088,:8089,:8090,:8091,:8092 &
PID2=$!
./out/shardsvr 3 :8088,:8089,:8090,:8091,:8092 &
PID3=$!
./out/shardsvr 4 :8088,:8089,:8090,:8091,:8092 &
PID4=$!

echo "Servers started with PIDs: $PID0, $PID1, $PID2, $PID3, $PID4"
echo "Press Ctrl+C to stop all servers"

# 等待 Ctrl+C
trap "kill $PID0 $PID1 $PID2 $PID3 $PID4; exit" SIGINT SIGTERM

wait
