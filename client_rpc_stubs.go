package main

import (
	"context"

	"log"
)

func GetPredecessorNode(ctx context.Context, address string) (*NodeAddr, error) {
	panic("not implmeneted")
}

func (node *Node) RCP_Ping(req *PingRequest, reply *PingResponse) error {
	// Fill reply pointer to send the data back

	log.Print("PINGED")

	if req.Message == "ping" {
		reply.Message = "ok"
	}

	return nil
}

func (node *Node) RCP_GetPredecessor(req *GetPredecessorRequest, reply *GetPredecessorResponse) error {
	// Fill reply pointer to send the data back

	log.Print("GetingPredecessor")

	if node.Predecessor != nil {
		reply.Node = NodePayload{NodeAddr: *node.Predecessor.Address}
	}

	return nil
}

func (node *Node) RCP_FindSuccessor(req *GetPredecessorRequest, reply *GetPredecessorResponse) error {
	if node.Successor != nil {
		reply.Node = NodePayload{NodeAddr: *node.Successor.Address}
	} else {
		reply.Node = NodePayload{NodeAddr: *node.Address}
	}
	return nil
}

func (node *Node) RCP_Notify(req *NotifyRequest, reply *NotifyResponse) error {

	if node.Predecessor == nil {
		node.Predecessor = &Node{Address: &req.Node.NodeAddr}
		return nil
	}

	/*
		// If req.Node.ID is between our predecessor and us, update
		if between(node.Predecessor.ID, big.NewInt(int(req.Node.ID)), node.ID, false) {
			node.Predecessor = &Node{Address: &req.Node.NodeAddr}
		} */
	return nil

}
