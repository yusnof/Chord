package main

import (
	"math/big"
	"strconv"
	"sync"
)



// Node represents a node in the Chord DHT
type Node struct {
	mu          sync.RWMutex
	ID          *big.Int
	Address     *NodeAddr
	Predecessor *Node
	Successor   *Node
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
	Node Node
}

type PingRequest struct {
	Message string
}

type PingResponse struct {
	Message string
}


type FindSuccessorRequest struct {
    ID string
}

type FindSuccessorResponse struct {
    Node Node
}


type NotifyRequest struct {
    Node Node
}

type NotifyResponse struct{}