package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"math/big"
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

	addr := FormatToString(n.IP, n.Port)

	if err := Call(addr, "find-succesor", &req, &res); err != nil {
		log.Printf("find-succesor: notify failed: %v", err)
		return nil
	}

	node := &Node{
		ID:   res.ID,
		IP:   res.IP,
		Port: res.Port,
	}

	n.PrintNode()
	return node

}

func checkForFileInBucket(file string, bucket map[string]string) bool {

	//traverse through the map
	for _, value := range bucket {

		//check if present value is equals to userValue
		if value == file {
			return true
		}
	}

	//if value not found return false
	return false
}

func (n *Node) Lookup(file string) (string, error) {

	//we look at file in the local storage
	n.mu.RLock()
	if checkForFileInBucket(file, n.Bucket) {
		log.Printf("Lookup: found %q locally", file)
		return n.Bucket[file], nil
	}
	n.mu.RUnlock()

	hashedFilename := hash(file)

	req := NodeInformationRequest{ID: hashedFilename}
	res := NodeInformationResponse{}

	addr := FormatToString(n.IP, n.Port)

	if err := Call(addr, "find-succesor", &req, &res); err != nil {
		return "", fmt.Errorf("Lookup: FindSuccessor failed: %w", err)
	}

	succAddr := FormatToString(res.IP, res.Port)
	fileReq := GetFileRequest{Filename: file}
	fileRes := GetFileResponse{}

	if err := Call(succAddr, "getFile", &fileReq, &fileRes); err != nil {
		return "", fmt.Errorf("Lookup: GetFile failed: %w", err)
	}

	if !fileRes.Found {
		return "", fmt.Errorf("Lookup: file %q not found", file)
	}

	return fileRes.Content, nil
	
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
	n.mu.RLock()
	succ := n.Successor
	selfID := n.ID
	n.mu.RUnlock()

	if succ == nil || selfID == nil {
		return
	}

	// Ask successor for its predecessor
	req := NodeInformationRequest{}
	res := NodeInformationResponse{}

	succAddr := FormatToString(succ.IP, succ.Port)

	if err := Call(succAddr, "getPred", &req, &res); err != nil {
		log.Printf("stabilize: notify failed: %v", err)
		return
	}

	// If successor's predecessor exists and is between us and successor, update
	if res.ID != nil && between(selfID, res.ID, succ.ID) {
		n.mu.Lock()
		n.Successor = &Node{
			ID:   res.ID,
			IP:   res.IP,
			Port: res.Port,
		}
		n.mu.Unlock()
	}

	// Notify current successor that we exist (so it updates its predecessor)
	notifyReq := NodeInformationRequest{
		ID:   n.ID,
		IP:   n.IP,
		Port: n.Port,
	}
	notifyReply := NodeInformationResponse{}

	n.mu.RLock()
	succAddr = FormatToString(n.Successor.IP, n.Successor.Port)
	n.mu.RUnlock()

	if err := Call(succAddr, "notify", &notifyReq, &notifyReply); err != nil {
		log.Printf("stabilize: notify failed: %v", err)
		return
	}

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
