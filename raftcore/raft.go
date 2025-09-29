package raftcore

import (
	
)


type Raft struct {
	NodeID string
	Role   RaftRole
	Term   int
	
}

type RaftRole int

const (
	RaftRoleFollower RaftRole = iota
	RaftRoleCandidate
	RaftRoleLeader
)

func MakeRaft (nodeID string,role RaftRole,term int)*Raft{
	return &Raft{
		NodeID:nodeID,
		Role:role,
		Term:term,
	}
} 

func main(){



}