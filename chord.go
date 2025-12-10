package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"math/big"
	"net/rpc"
)

const (
	keySize = sha1.Size * 8
	m       = 3
)

func hash(elt string) *big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(elt))
	return new(big.Int).SetBytes(hasher.Sum(nil))
}

// returns true if elt is between start and end, accounting for the right
// if inclusive is true, it can match the end
func between(start, elt, end *big.Int) bool {
	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (elt.Cmp(end) == 0)
	}
}

func (n *Node) Create() {
	n.mu.Lock()
	log.Print("Creating")
	n.Predecessor = nil // not need to hold this

	n.ID = hash(FormatToString(n.IP, n.Port))

	n.Successor = n
	n.mu.Unlock()
}

func (n *Node) Join(node *Node) {
	n.mu.Lock()
	log.Print("Joining")
	n.ID = hash(FormatToString(n.IP, n.Port))

	//maybe fix  later
	n.Predecessor = node

	n.mu.Unlock()

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
		log.Printf("%v : Empty Predecessor", FormatToString(n.IP, n.Port))
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

	client, err := rpc.DialHTTP("tcp", FormatToString(n.IP, n.Port))
	log.Printf("Pinging my pred: %v", FormatToString(n.IP, n.Port))
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
