package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// PingNode sends a ping to another node
func PingNode(ctx context.Context, address string) error {
	address = resolveAddress(address)
	log.Printf("Pinging IP: %s", address)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	return nil
}



func GetPredecessorNode(ctx context.Context, address string) (*NodeAddr, error) {
	panic("not implmeneted")
}

func (node *Node) Ping(req *PingRequest, reply *PingResponse) error {
	// Fill reply pointer to send the data back

	log.Print("PINGED")

	if req.Message == "ping" {
		reply.Message = "ok"
	}

	return nil
}
