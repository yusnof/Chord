package main

import (
	"log"
	"math/big"
	"strconv"
	"sync"
)

// Node represents a node in the Chord DHT
type Node struct {
	mu sync.RWMutex
	ID *big.Int

	IP   string
	Port int

	Predecessor *Node
	Successor   *Node
	FingerTable []*Node

	Bucket map[string]string
}

func (n *Node) PrintNode() {
	predAddr := "Nil"
	if n.Predecessor != nil {
		predAddr = FormatToString(n.Predecessor.IP, n.Predecessor.Port)
	}
	succAddr := "Nil"
	if n.Successor != nil {
		succAddr = FormatToString(n.Successor.IP, n.Successor.Port)
	}
	log.Printf("Node %s | ID: %v | Pred: %s | Succ: %s",
		FormatToString(n.IP, n.Port), n.ID, predAddr, succAddr)
}

func FormatToString(ip string, port int) string {
	return ip + ":" + strconv.Itoa(port)
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

type PingRequest struct {
	Message string
}

type PingResponse struct {
	Message string
}

type NodeInformationRequest struct {
	ID   *big.Int
	IP   string
	Port int
}

type NodeInformationResponse struct {
	ID   *big.Int
	IP   string
	Port int
}
