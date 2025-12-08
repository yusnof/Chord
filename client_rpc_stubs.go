package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
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

	client := NewChordClient(conn)
	_, err = client.Ping(ctx, PingRequest{})

	if err != nil {
		return fmt.Errorf("ping failed: %v", err)
	}

	return nil
}

type chordClient struct {
	conn *grpc.ClientConn
}

// GetPredecessor implements ChordClient.
func (c *chordClient) GetPredecessor(ctx context.Context, in GetPredecessorRequest, opts ...grpc.CallOption) (GetPredecessorResponse, error) {
	panic("unimplemented")
}

// Ping implements ChordClient.
func (c *chordClient) Ping(ctx context.Context, in PingRequest, opts ...grpc.CallOption) (PingResponse, error) {
	panic("unimplemented")
}

// Constructor
func NewChordClient(conn *grpc.ClientConn) ChordClient {
	return &chordClient{conn: conn}
}

func GetPredecessorNode(ctx context.Context, address string) (*NodeAddr, error) {
	address = resolveAddress(address)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	client := NewChordClient(conn)

	resp, err := client.GetPredecessor(ctx, GetPredecessorRequest{})

	if err != nil {
		return nil, err
	}

	return &NodeAddr{IP: resp.Node.IP, Port: int(resp.Node.Port)}, nil
}
