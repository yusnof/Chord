package main

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log"
	"math/big"
	"net/rpc"
)

const (
	successorListSize = 3
	keySize           = sha1.Size * 8
	maxLookupSteps    = 32
)

var (
	two     = big.NewInt(2)
	hashMod = new(big.Int).Exp(big.NewInt(2), big.NewInt(keySize), nil)
)

func (n *Node) Create() {
	n.mu.Lock()
	log.Print("Creating")
	n.Predecessor = nil // not need to hold this
	n.Successors = make([]*NodeAddr, successorListSize)
	n.Successors[0] = n.Address
	n.mu.Unlock()
}

func (n *Node) Join(node_addr *NodeAddr) {
	n.mu.Lock()
	log.Print("Joining")
	//maybe fix  later
	n.Predecessor = node_addr
	n.Successors = make([]*NodeAddr, successorListSize)

	n.mu.Unlock()

}

func GetAllKeyValues(ctx context.Context, s string) (any, any) {
	panic("unimplemented")
}

func Lookup() any {
	panic("unimplemented")
}
func StoreFile() any {
	panic("unimplemented")
}
func PrintState() any {
	panic("unimplemented")
}

//TODO the node should ping the node before to know that its alive,

// of resose.
func (n *Node) checkPredecessor() {
	n.mu.Lock()
	defer n.mu.Unlock()

	//we have no predecessor, exit if
	if n.Predecessor == nil {
		log.Printf("%v : Empty Predecessor", ToString(n.Address))
		return
	} else {

		err := n.Call(ToString(n.Predecessor), "PING", PingRequest{}, PingResponse{})

		if err != nil {
			log.Print("Error in checkPredecessor:", err)
			n.Predecessor = nil
		}

	}

}

func (n *Node) stabilize() {
	// TODO: Student will implement this

	//n.call(ToString(n.Predecessor), "GET_Predecessor", PingRequest{}, PingResponse{})

}

func (n *Node) fixFingers(nextFinger int) int {
	// TODO: Student will implement this
	nextFinger++
	if nextFinger > keySize {
		nextFinger = 1
	}
	return nextFinger
}

// will be used in order to call the other nodes.
func (n *Node) Call(address string, method string, request interface{}, reply interface{}) error {

	if method == "PING" {
		req := PingRequest{}
		res := PingResponse{}

		req.Message = "ping"

		client, err := rpc.DialHTTP("tcp", address)
		log.Printf("Pinging my pred: %w", address)
		if err != nil {
			log.Fatal("dialing:", err)
		}
		err = client.Call("Node.Ping", &req, &res)
		if err != nil {
			return fmt.Errorf("ping failed: %w", err)
		}
		if(res.Message != "ok"){
			return fmt.Errorf("ping failed:", "withoutok")

		}

	}

	return nil
}

func (n *Node) GetPredecessor(ctx context.Context, req GetPredecessorRequest) (GetPredecessorResponse, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.Predecessor == nil {
		return GetPredecessorResponse{Node: *n.Address}, nil
	}
	return GetPredecessorResponse{
		Node: NodeAddr{
			IP:   n.Predecessor.IP,
			Port: int(n.Predecessor.Port),
		},
	}, nil
}
