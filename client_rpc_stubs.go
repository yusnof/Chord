package main

import (
	"fmt"
	"log"
	"net/rpc"
)

func (node *Node) RCP_Ping(req *PingRequest, reply *PingResponse) error {
	// Fill reply pointer to send the data back

	log.Print("Recived a Ping")

	if req.Message == "ping" {
		reply.Message = "ok"
	}

	return nil
}

func (node *Node) RCP_FindSuccessor(req *NodeInformationRequest, reply *NodeInformationResponse) error {
	// Fill reply pointer to send the data back

	log.Print("Recived a request for finding a successor")

	node.mu.RLock()
	succ := node.Successor
	selfID := node.ID
	node.mu.RUnlock()

	
		if succ == nil || selfID == nil {
			log.Print("shit")
			reply.ID = node.ID
			reply.IP = node.IP
			reply.Port = node.Port
			return nil
		} 

	// this is the case of the first node
	if succ.ID == selfID {
		node.mu.Lock()
		node.Successor = &Node{
			ID:   req.ID,
			IP:   req.IP,
			Port: req.Port,
		}
		node.mu.Unlock()

		node.PrintNode()

		// Notify the joining node that we (the first node) are its predecessor
		notifyReq := NodeInformationRequest{
			ID:   node.ID,
			IP:   node.IP,
			Port: node.Port,
		}

		notifyReply := NodeInformationResponse{}

		Call(FormatToString(req.IP, req.Port), "notify", &notifyReq, &notifyReply)

		node.PrintNode()

		return nil

	}

	//in the case of more than one node
	if between(selfID, req.ID, succ.ID) {
		reply.ID = succ.ID
		reply.IP = succ.IP
		reply.Port = succ.Port
		return nil
	}

	succAddr := FormatToString(succ.IP, succ.Port)
	client, err := rpc.DialHTTP("tcp", succAddr)
	if err != nil {
		// fallback: return our successor if dialing fails
		log.Printf("RCP_FindSuccessor: dialing successor %s failed: %v", succAddr, err)
		reply.ID = succ.ID
		reply.IP = succ.IP
		reply.Port = succ.Port
		return nil
	}
	defer client.Close()

	// Forward same request
	if err := client.Call("Node.RCP_FindSuccessor", req, reply); err != nil {
		log.Printf("RCP_FindSuccessor: forwarding failed: %v", err)
		// fallback: return our successor
		reply.ID = succ.ID
		reply.IP = succ.IP
		reply.Port = succ.Port
		return nil
	}

	return nil
}

func Call(address string, rpcMethod string, req interface{}, reply interface{}) error {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		return fmt.Errorf("Call: dialing %s: %w", address, err)
	}
	defer client.Close()

	switch rpcMethod {

	case "getPred":
		if err := client.Call("Node.RCP_GetPredecessor", req, reply); err != nil {
			log.Printf("stabilize: GetPredecessor RPC failed: %v", err)
			return err
		}

	case "notify":
		if err := client.Call("Node.RCP_Notify", req, reply); err != nil {
			log.Printf("failes notify: %v", err)
			return err
		}
	case "getFile":
		if err := client.Call("Node.RCP_GetFile", req, reply); err != nil {
			return fmt.Errorf("Call: getFile RPC failed: %w", err)
		}

	case "ping":
		log.Printf("Pinging my pred: %v", address)
		err = client.Call("Node.RCP_Ping", req, reply)
		if err != nil {
			return fmt.Errorf("ping failed: %w", err)
		}

	case "find-succesor":
		if err := client.Call("Node.RCP_FindSuccessor", req, reply); err != nil {
			log.Printf("find_successor RPC failed: %v", err)
			return fmt.Errorf("finding succeser failed: %w", err)
		}
	}

	return nil

}

func (node *Node) RCP_Notify(req *NodeInformationRequest, reply *NodeInformationResponse) error {
	log.Print("GetingNotified")

	node.Predecessor = &Node{
		ID:   req.ID,
		IP:   req.IP,
		Port: req.Port,
	}

	return nil

}

func (node *Node) RCP_GetPredecessor(req *NodeInformationRequest, reply *NodeInformationResponse) error {
	log.Print("GettingPredecessor")

	node.mu.RLock()
	defer node.mu.RUnlock()

	if node.Predecessor != nil {
		reply.ID = node.Predecessor.ID
		reply.IP = node.Predecessor.IP
		reply.Port = node.Predecessor.Port
	}
	
	return nil
}

func (node *Node) RCP_GetFile(req *GetFileRequest, reply *GetFileResponse) error {
	log.Print("RCP_GetFile")

	node.mu.RLock()
	defer node.mu.RUnlock()

	if node.Bucket == nil {
		reply.Found = false
		return nil
	}

	if content, ok := node.Bucket[req.Filename]; ok {
		reply.Found = true
		reply.Content = content
		return nil
	}

	reply.Found = false
	return nil
}


func (node *Node) RCP_StoreFile(req *StoreFileRequest, reply *StoreFileResponse) error {
    log.Printf("RCP_StoreFile: storing %s", req.Filename)

    node.mu.Lock()
    defer node.mu.Unlock()

    if node.Bucket == nil {
        node.Bucket = make(map[string]string)
    }

    
    node.Bucket[req.Filename] = string(req.Content)

    reply.Success = true
    reply.Message = "stored"
    return nil
}