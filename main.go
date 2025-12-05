package main

import (
	"bufio"
	pb "chord/protocol"
	"io"
	"os"

	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
)

var (
	IP_addr         string
	Port            int
	joinAddr        string
	joinPort        int
	ts              int
	tff             int
	tcp             int
	r               int
	flag_first_node bool
	i               string
	localaddress    string // TODO to be removed
)

func init() {
	flag.StringVar(&IP_addr, "a", "127.0.0.1", "Chord IP Address")
	flag.IntVar(&Port, "p", 1234, "The port that the Chord client will bind to and listen on. Represented as a base-10 integer.")

	flag.StringVar(&joinAddr, "ja", "", "The IP address of the machine running a Chord node. The Chord client will join this node’s ring")

	flag.IntVar(&joinPort, "jp", 0, "The port that an existing Chord node is bound to and listening on.")
	flag.IntVar(&ts, "ts", 3000, "The time in milliseconds between invocations of ‘stabilize’. Represented as a base-10 integer.")
	flag.IntVar(&tff, "tff", 1000, "The time in milliseconds between invocations of ‘fix fingers’.")
	flag.IntVar(&tcp, "tcp", 3000, "The time in milliseconds between invocations of check predecessor’")
	flag.IntVar(&r, "r", 4, "The number of successors maintained by the Chord client. ")
	flag.StringVar(&i, "i", "", "he identifier (ID) assigned to the Chord client which will override the ID computed by the SHA1 sum of the client’s IP address and port number.")

	flag.Parse()

	validateInput()

}

func main() {

	node, _ := StartServer()

	go RunShell(node)
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
				fmt.Printf("Ping failed: %v\n", err)
			} else {
				fmt.Println("Ping successful")
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

func StartServer() (*Node, error) {

	address := fmt.Sprintf(":%d", Port)

	log.Print("heloo")

	node := &Node{
		Address:     address,
		FingerTable: make([]string, keySize+1),
		Predecessor: "", //TODO, to chech were to update this
		Successors:  nil,
		Bucket:      make(map[string]string),
	}
	//TOD add more logic here

	// Start listening for RPC calls
	grpcServer := grpc.NewServer()

	pb.RegisterChordServer(grpcServer, node)

	lis, err := net.Listen("tcp", node.Address)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		log.Printf("Starting Chord node server on %s", node.Address)

	}()
	if flag_first_node {
		node.Successors = []string{node.Address}
	} else {
		// node.Successors = []string{nprime} TODO not sure about this one
	}

	go func() {
		nextFinger := 0
		for {
			time.Sleep(time.Millisecond * time.Duration(ts))
			node.stabilize()

			time.Sleep(time.Millisecond * time.Duration(tff))
			nextFinger = node.fixFingers(nextFinger)

			time.Sleep(time.Millisecond * time.Duration(tcp))
			node.checkPredecessor()
		}
	}()

	return node, nil
}

func validateInput() {
	flag_first_node = true

	if IP_addr == "" || net.ParseIP(IP_addr).To4() == nil {
		log.Fatal("-a must be a valid IPv4 address, e.g. 128.8.126.63")
	}
	if Port < 1024 || Port > 65535 {
		log.Fatal("-p must be in range 1024–65535")
	}

	if (joinAddr == "" && joinPort != 0) || (joinAddr != "" && joinPort == 0) {
		log.Fatal("--ja and --jp must be given together, or neither")
	}

	if joinAddr != "" && joinPort != 0 {
		flag_first_node = false
		if net.ParseIP(joinAddr).To4() == nil {
			log.Fatal("--ja must be a valid IPv4 address")
		}
		if joinPort < 1024 || joinPort > 65535 {
			log.Fatal("--jp must be in range 1024–65535")
		}
	}

	if ts <= 1 || ts >= 60000 {
		log.Fatal("Ts must be specified, or between 1 and 60k")

	}
	if tff <= 1 || tff >= 60000 {
		log.Fatal("Tff must be specified, or between 1 and 60k")

	}
	if tcp <= 1 || tcp >= 60000 {
		log.Fatal("tcp must be specified, or between 1 and 60k")

	}
	if r <= 1 || r >= 32 {
		log.Fatal("R must be specified, or between 1 and 32")
	}
	if i != "" {
		if len(i) != 40 || !isHex(i) {
			log.Fatal("-i must be 40 hex chars (0-9a-fA-F)")
		}

	}
}

func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'f') ||
			(c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func resolveAddress(address string) string {
	if strings.HasPrefix(address, ":") {
		return net.JoinHostPort(localaddress, address[1:])
	} else if !strings.Contains(address, ":") {
		return net.JoinHostPort(address, defaultPort)
	}
	return address
}
