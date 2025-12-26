package raftcore

import (
"testing"
)

func TestMakeRaftClient(t *testing.T) {
// 定义测试参数
addr := "127.0.0.1:8080"
id := int64(100)

// 调用被测函数
client := MakeRaftClient(addr, id)

// 验证结果不为 nil
if client == nil {
t.Fatal("MakeRaftClient returned nil")
}

// 验证 ID 是否正确设置
if got := client.GetId(); got != id {
t.Errorf("GetId() = %v, want %v", got, id)
}

// 验证 MessageServiceClient 是否已初始化
if client.GetMessageService() == nil {
t.Error("GetMessageService() returned nil")
}

// 验证内部字段 (由于是同一个包，可以访问私有字段)
if client.addr != addr {
t.Errorf("client.addr = %v, want %v", client.addr, addr)
}

if client.conn == nil {
t.Error("client.conn is nil")
}
}
