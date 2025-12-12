package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
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

	n.Bucket = make(map[string]string, 3)

	n.PrintNode()
	n.Successor = n
	n.mu.Unlock()
}

func (n *Node) Join(npnode *Node) {
	n.mu.Lock()
	log.Print("Joining")
	n.ID = hash(FormatToString(n.IP, n.Port))

	//maybe fix  later
	n.Predecessor = nil
	n.Successor = npnode.find_successor(n.ID, n.IP, n.Port)

	n.Bucket = make(map[string]string, 3)

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

	// maybe this is an ovekill look dir from the stored info
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
func (n *Node) PrintState() {

	n.mu.RLock()
	defer n.mu.RUnlock()

	selfAddr := FormatToString(n.IP, n.Port)
	selfID := "<nil>"
	if n.ID != nil {
		selfID = n.ID.String()
	}

	predAddr, predID := "Nil", "<nil>"
	if n.Predecessor != nil {
		predAddr = FormatToString(n.Predecessor.IP, n.Predecessor.Port)
		if n.Predecessor.ID != nil {
			predID = n.Predecessor.ID.String()
		}
	}

	succAddr, succID := "Nil", "<nil>"
	if n.Successor != nil {
		succAddr = FormatToString(n.Successor.IP, n.Successor.Port)
		if n.Successor.ID != nil {
			succID = n.Successor.ID.String()
		}
	}

	// Print header
	fmt.Printf("Node %s\n", selfAddr)
	fmt.Printf("  ID        : %s\n", selfID)
	fmt.Printf("  Predecessor: %s (ID=%s)\n", predAddr, predID)
	fmt.Printf("  Successor  : %s (ID=%s)\n", succAddr, succID)

	/*
		// Finger table summary (if present)
		if n.FingerTable != nil {
			fmt.Printf("  Fingers (%d):\n", len(n.FingerTable))
			for i, f := range n.FingerTable {
				if f == nil {
					fmt.Printf("    [%2d] nil\n", i)
				} else {
					fid := "<nil>"
					if f.ID != nil {
						fid = f.ID.String()
					}
					fmt.Printf("    [%2d] %s (ID=%s)\n", i, FormatToString(f.IP, f.Port), fid)
				}
			}
		} else {
			fmt.Println("  Fingers: nil")
		} */

	// Bucket info
	if n.Bucket != nil {
		fmt.Printf("  Bucket: %d entries\n", len(n.Bucket))
	} else {
		fmt.Println("  Bucket: nil")
	}

}

// TODO the node should ping the node before to know that its alive,
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

func (n *Node) StoreFile(localPath string, password string) (string, error) {
	// read file bytes
	data, err := os.ReadFile(localPath)
	if err != nil {
		return "", fmt.Errorf("StoreFile: read file: %w", err)
	}
	
	filename := filepath.Base(localPath)

	// 1) check if we already have it locally
	n.mu.RLock()
	if n.Bucket != nil {
		if _, ok := n.Bucket[filename]; ok {
			n.mu.RUnlock()
			log.Printf("StoreFile: %s already local", filename)
			return "", nil
		}
	}
	n.mu.RUnlock()

	// 2) find successor responsible for file key
	keyID := hash(filename)
	req := NodeInformationRequest{ID: keyID}
	res := NodeInformationResponse{}
	if err := Call(FormatToString(n.IP, n.Port), "find-succesor", &req, &res); err != nil {
		return "", fmt.Errorf("StoreFile: find successor failed: %w", err)
	}

	succAddr := FormatToString(res.IP, res.Port)

	// 3) send store RPC to responsible node
	storeReq := StoreFileRequest{Filename: filename, Content: (data)}
	storeRes := StoreFileResponse{}

	if err := Call(succAddr, "storeFile", &storeReq, &storeRes); err != nil {
		return "", fmt.Errorf("StoreFile: remote store failed: %w", err)
	}
	if !storeRes.Success {
		return "", fmt.Errorf("StoreFile: remote store error: %s", storeRes.Message)
	}

	return succAddr, nil
}
