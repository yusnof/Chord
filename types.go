package main

import (
	pb "chord/protocol"
	"strconv"
	"sync"
)

// Node represents a node in the Chord DHT
type Node struct {
	pb.UnimplementedChordServer
	mu          sync.RWMutex
	Address     *NodeAddr
	Predecessor *NodeAddr 
	Successors  []*Node
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
func (n *Node) toString() string {
	return "node: IP:" + n.Address.IP + ":" + strconv.Itoa(n.Address.Port) + "\n" + "Predecessor:" + "\n"
}

type Config struct {
    IPAddr   string
    Port     int
    JoinAddr string
    JoinPort int
    TS       int
    TFF      int
    TCP      int
    R        int
    I       string
    Flag_first_node bool
}