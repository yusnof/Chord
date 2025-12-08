package main

import (
	"context"
	"strconv"
	"sync"

	"google.golang.org/grpc"
)

type ChordClient interface {
    Ping(ctx context.Context, in PingRequest, opts ...grpc.CallOption) (PingResponse, error)
    GetPredecessor(ctx context.Context, in GetPredecessorRequest, opts ...grpc.CallOption) (GetPredecessorResponse, error)
}

// Node represents a node in the Chord DHT
type Node struct {
	mu          sync.RWMutex
	Address     *NodeAddr
	Predecessor *NodeAddr
	Successors  []*NodeAddr
	FingerTable []string

	Bucket map[string]string
}

type NodeAddr struct {
	IP   string
	Port int
}

func ToString(node_addr *NodeAddr) string {
	return node_addr.IP + ":" + strconv.Itoa(node_addr.Port)
}

type Config struct {
	IPAddr          string
	Port            int
	JoinAddr        string
	JoinPort        int
	TS              int
	TFF             int
	TCP             int
	R               int
	I               string
	Flag_first_node bool
}

type GetPredecessorRequest struct{}

type GetPredecessorResponse struct {
	Node NodeAddr
}

type PingRequest struct{}

type PingResponse struct{}