package main

import (
	"bufio"
	pb "chord/protocol"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
)

var (
	node_addr    IPandPortAddr
	localaddress string // TODO to be removed
)



func main() {

	// will be loging into a file in order to not block the CLI 
	logFile := LogerConfigurationSetup()
    defer logFile.Close()  
    defer logFile.Sync()

	log.Println("Chord: Logging started")

	cfg := Loadconfig()

	node := StartServer(cfg)

	RunShell(node)
}
func RunShell(node *Node) {
	reader := bufio.NewReader(os.Stdin)
	for {

		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nExiting...")
				return
			}
			fmt.Println("Error reading input:", err)
			continue
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		//defer cancel()

		switch parts[0] {
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  help              - Show this help message")
			fmt.Println("  Lookup     -takes as input the name of a file to be searched (e.g., “Hello.txt”).")
			fmt.Println("  StoreFile  -takes the location of a file on a local disk, then performs a lookup to find the Chord node to store the file at")
			fmt.Println("  PrintState - requires no input. The Chord client outputs its local state information at the current time")

			fmt.Println("  quit              - Exit the program")

		case "Lookup":
			if len(parts) < 2 {
				fmt.Println("Usage: Lookup <File Name>")
				continue
			}
			err := Lookup()

			if err != nil {
				fmt.Printf("Lookup failed: %v\n", err)
			} else {
				fmt.Println("Lookup successful")
			}

		case "StoreFile":
			//TODO
		case "PrintState":
			//TODO

		case "quit":
			fmt.Println("Exiting...")
			return

		default:
			fmt.Println("Unknown command. Type 'help' for available commands.")

		}

	}
}

func StartServer(cfg Config) *Node {

	node := &Node{
		Address:     My_IP_tostring(),
		FingerTable: make([]string, keySize+1),
		Successors:  nil,
		Bucket:      make(map[string]string),
	}
	//TOD add more logic here

	// Start listening for RPC calls
	grpcServer := grpc.NewServer()

	pb.RegisterChordServer(grpcServer, node)

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Port))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("Starting Chord node server on %s", strconv.Itoa(cfg.Port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	if cfg.Flag_first_node {
		node.Create()
	} else {
		node.Join()
	}

	node.toString()

	go func() {
		nextFinger := 0
		for {
			time.Sleep(time.Millisecond * time.Duration(cfg.TS))
			node.stabilize()

			time.Sleep(time.Millisecond * time.Duration(cfg.TFF))
			nextFinger = node.fixFingers(nextFinger)

			time.Sleep(time.Millisecond * time.Duration(cfg.TCP))
			node.checkPredecessor()
		}
	}()

	return node
}

func resolveAddress(address string) string {
	if strings.HasPrefix(address, ":") {
		return net.JoinHostPort(localaddress, address[1:])
	} else if !strings.Contains(address, ":") {
		return net.JoinHostPort(address, strconv.Itoa(node_addr.Port))
	}
	return address
}
