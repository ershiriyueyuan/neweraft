package raftcore

import (
"context"
"testing"
"time"

"google.golang.org/grpc"
pb "neweraft/raftpb"
)

// MockMessageServiceClient implements pb.MessageServiceClient for testing
type MockMessageServiceClient struct {
voteResponse        *pb.VoteResponse
appendEntryResponse *pb.AppendEntryResponse
err                 error
}

func (m *MockMessageServiceClient) RequestVote(ctx context.Context, in *pb.VoteRequest, opts ...grpc.CallOption) (*pb.VoteResponse, error) {
return m.voteResponse, m.err
}

func (m *MockMessageServiceClient) AppendEntry(ctx context.Context, in *pb.AppendEntryRequest, opts ...grpc.CallOption) (*pb.AppendEntryResponse, error) {
return m.appendEntryResponse, m.err
}

func TestMakeRaft(t *testing.T) {
peers := []*RaftClient{}
raft := MakeRaft("127.0.0.1:8000", 1, peers)

if raft.id != 1 {
t.Errorf("expected id 1, got %d", raft.id)
}
if raft.role != RaftFollower {
t.Errorf("expected role Follower, got %v", raft.role)
}
if raft.curTerm != 0 {
t.Errorf("expected term 0, got %d", raft.curTerm)
}

raft.deadIf = true
}

func TestHandleRequestVote(t *testing.T) {
peers := []*RaftClient{}
raft := MakeRaft("127.0.0.1:8000", 1, peers)
raft.deadIf = true 

req := &pb.VoteRequest{
CurTerm: 1,
SefId:   2,
}
res := &pb.VoteResponse{}

raft.HandleRequestVote(req, res)

if !res.VoteGranted {
t.Error("expected vote granted for higher term")
}

req.CurTerm = 0
res.VoteGranted = false
raft.curTerm = 1
raft.HandleRequestVote(req, res)

if res.VoteGranted {
t.Error("expected vote denied for lower term")
}
}

func TestElection(t *testing.T) {
mockClient := &MockMessageServiceClient{
voteResponse: &pb.VoteResponse{VoteGranted: true},
}

// Simulate 3 nodes cluster
peer1 := &RaftClient{id: 1, MessageServiceClient: mockClient} // Self
peer2 := &RaftClient{id: 2, MessageServiceClient: mockClient}
peer3 := &RaftClient{id: 3, MessageServiceClient: mockClient}
peers := []*RaftClient{peer1, peer2, peer3}

raft := MakeRaft("127.0.0.1:8000", 1, peers)
raft.deadIf = true 

// Manually trigger election
raft.election()

// Wait for goroutines
time.Sleep(100 * time.Millisecond)

raft.mu.Lock()
defer raft.mu.Unlock()

if raft.role != RaftLeader {
t.Errorf("expected to become leader, got %v", raft.role)
}
if raft.curTerm != 1 {
t.Errorf("expected term incremented to 1, got %d", raft.curTerm)
}
}

func TestSwitchRole(t *testing.T) {
peers := []*RaftClient{}
raft := MakeRaft("127.0.0.1:8000", 1, peers)
raft.deadIf = true

raft.switchRole(RaftCandidate)
if raft.role != RaftCandidate {
t.Errorf("expected role Candidate, got %v", raft.role)
}

raft.switchRole(RaftLeader)
if raft.role != RaftLeader {
t.Errorf("expected role Leader, got %v", raft.role)
}
}
