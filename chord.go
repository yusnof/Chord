package main

import (
	"crypto/sha1"
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

	n.PrintNode()
	n.Successor = n
	n.mu.Unlock()
}

func (n *Node) Join(npnode *Node) {
	n.mu.Lock()
	log.Print("Joining")
	n.ID = hash(FormatToString(n.IP, n.Port))

	//npnode.ID = hash(FormatToString(n.IP, n.Port))

	//maybe fix  later
	n.Predecessor = nil

	n.Successor = npnode.find_successor(n.ID, n.IP, n.Port)

	n.PrintNode()
	n.mu.Unlock()

}

func (n *Node) find_successor(id *big.Int, ip string, port int) *Node {

	log.Print("finding a succesor")

	req := NodeInformationRequest{
		ID:   id,
		IP:   ip,
		Port: port,
	}
	res := NodeInformationResponse{}

	client, err := rpc.DialHTTP("tcp", FormatToString(n.IP, n.Port))

	log.Printf("finding_successor from: %v", FormatToString(n.IP, n.Port))

	if err != nil {
		log.Fatal("dialing:", err)
	}
	if err := client.Call("Node.RCP_FindSuccessor", &req, &res); err != nil {
		log.Printf("find_successor RPC failed: %v", err)
		return nil
	}
	defer client.Close()

	node := &Node{
		ID:   res.ID,
		IP:   res.IP,
		Port: res.Port,
	}

	log.Println("here")
	n.PrintNode()
	return node

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

	Call(FormatToString(n.Predecessor.IP, n.Predecessor.Port), "ping", &req, &res)

	return nil
}
