package shardkvserver

import (
    "testing"
)

func TestMakeShardServer(t *testing.T) {
    // 1. 准备模拟的节点地址映射
    peersMap := map[int]string{
        0: "localhost:5000",
        1: "localhost:5001",
        2: "localhost:5002",
    }

    // 2. 测试当前节点 ID
    me := int64(0)

    // 3. 尝试创建 ShardServer
    // 注意：这会启动 Raft 实例，可能会尝试连接其他节点（虽然这里没有真正启动 gRPC 服务）
    server := MakeShardServer(peersMap, me)

    // 4. 验证服务器是否成功创建
    if server == nil {
        t.Fatal("MakeShardServer returned nil")
    }

    // 验证内部结构是否初始化 (由于字段是私有的，我们只能做基本检查或通过反射，
    // 但只要上面没有 panic 且不为 nil，通常表示初始化流程走通了)
    t.Logf("Successfully created ShardServer for node %d", me)
}// filepath: /home/yuan/neweraft/shardkvserver/shardkv_server_test.go
