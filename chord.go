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
	keySize = sha1.Size * 8
)

func hash(elt string) *big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(elt))
	return new(big.Int).SetBytes(hasher.Sum(nil))
}

// returns true if elt is between start and end, accounting for the right
// if inclusive is true, it can match the end
func between(start, elt, end *big.Int, inclusive bool) bool {
	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}
}

func (n *Node) Create() {
	n.mu.Lock()
	log.Print("Creating")
	n.Predecessor = nil // not need to hold this
	n.ID = hash(ToString(n.Address))
	n.Successor = n
	n.mu.Unlock()
}

func (n *Node) Join(node *Node) {
	n.mu.Lock()
	log.Print("Joining")
	n.ID = hash(ToString(n.Address))
	//maybe fix  later
	//n.Predecessor = node

	n.Successor = node

	n.FindSuccessor(node)

	n.mu.Unlock()

}

func (n *Node) FindSuccessor(node *Node) error {

	req := FindSuccessorRequest{ID: n.ID.String()}
	res := FindSuccessorResponse{}

	client, err := rpc.DialHTTP("tcp", ToString(node.Address))
	if err != nil {
		log.Printf("Join: dialing bootstrap %s failed: %v", ToString(node.Address), err)
		return err
	}
	err = client.Call("Node.RCP_FindSuccessor", &req, &res)
	if err != nil {
		log.Printf("Join: FindSuccessor RPC failed: %v", err)
		return err
	}

	n.mu.Lock()
	n.Successor = &res.Node
	n.mu.Unlock()

	n.Notify()

	return nil
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

		err := n.Ping()

		if err != nil {
			log.Print("Error in prev node, maybe dead:", err)
			n.Predecessor = nil
		}

	}

}

func (n *Node) stabilize() {
	// TODO: Student will implement this

	/*
		log.Print("Stabilizing")
		x, err := n.Successor.GetPredecessor()

		if err != nil {
			log.Printf("Error: %w", err)
			return
		}

		value := between(x.ID, n.ID, n.Successor.ID, false)

		if value {
			n.Successor = x
		}

		n.Successor.Notify() */

}

func (n *Node) fixFingers(nextFinger int) int {
	// TODO: Student will implement this
	nextFinger++
	if nextFinger > keySize {
		nextFinger = 1
	}
	return nextFinger
}

func (n *Node) Ping() error {

	req := PingRequest{}
	res := PingResponse{}

	req.Message = "ping"

	client, err := rpc.DialHTTP("tcp", ToString(n.Predecessor.Address))
	log.Printf("Pinging my pred: %v", ToString(n.Predecessor.Address))
	if err != nil {
		log.Fatal("dialing:", err)
	}
	err = client.Call("Node.RCP_Ping", &req, &res)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	if res.Message != "ok" {
		return fmt.Errorf("ping failed:%v", 1)
	}
	return nil
}

func (node *Node) GetPredecessor() (*Node, error) {
	if node.Predecessor != nil {
		req := GetPredecessorRequest{}
		res := GetPredecessorResponse{}

		client, err := rpc.DialHTTP("tcp", ToString(node.Predecessor.Address))

		log.Printf("Geting Pred: %v", ToString(node.Predecessor.Address))

		if err != nil {
			log.Fatal("dialing:", err)
		}
		err = client.Call("Node.RCP_GetPredecessor", &req, &res)

		if err != nil {
			return &Node{}, fmt.Errorf("failed: %w", err)
		}

		return &res.Node, nil
	} else {
		return &Node{}, fmt.Errorf("shit")
	}
}

func (n *Node) Notify() {
	n.mu.RLock()
	succ := n.Successor
	n.mu.RUnlock()

	if succ == nil {
		return
	}

	req := NotifyRequest{Node: *n}
	res := NotifyResponse{}

	client, err := rpc.DialHTTP("tcp", ToString(succ.Address))
	if err != nil {
		log.Printf("Notify: dialing successor %s failed: %v", ToString(succ.Address), err)
		return
	}
	if err := client.Call("Node.RCP_Notify", &req, &res); err != nil {
		log.Printf("Notify: RPC to successor failed: %v", err)
	}

}
