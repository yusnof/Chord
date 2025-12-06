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
	Address     string
	Predecessor *Node
	Successors  []*Node
	FingerTable []string

	Bucket map[string]string
}
type IPandPortAddr struct {
	IP   string
	Port int
}

func My_IP_tostring() string {
	return node_addr.IP + ":" + strconv.Itoa(node_addr.Port)
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