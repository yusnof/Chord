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
	flag_first_node bool
	localaddress    string // TODO to be removed
)

func loadconfig() Config {
	cfg := Config{}

	flag.StringVar(&cfg.IPAddr, "a", "127.0.0.1", "Chord IP Address")
	flag.IntVar(&cfg.Port, "p", 1234, "The port that the Chord client will bind to and listen on. Represented as a base-10 integer.")

	flag.StringVar(&cfg.JoinAddr, "ja", "", "The IP address of the machine running a Chord node. The Chord client will join this node’s ring")

	flag.IntVar(&cfg.JoinPort, "jp", 0, "The port that an existing Chord node is bound to and listening on.")
	flag.IntVar(&cfg.TS, "ts", 3000, "The time in milliseconds between invocations of ‘stabilize’. Represented as a base-10 integer.")
	flag.IntVar(&cfg.TFF, "tff", 1000, "The time in milliseconds between invocations of ‘fix fingers’.")
	flag.IntVar(&cfg.TCP, "tcp", 3000, "The time in milliseconds between invocations of check predecessor’")
	flag.IntVar(&cfg.R, "r", 4, "The number of successors maintained by the Chord client. ")
	flag.StringVar(&cfg.I, "i", "", "he identifier (ID) assigned to the Chord client which will override the ID computed by the SHA1 sum of the client’s IP address and port number.")

	flag.Parse()

	validateInput(cfg)

	return cfg

}

func main() {

	cfg := loadconfig()

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

func StartServer(cfg Config) *Node {

	address := fmt.Sprintf(":%d", cfg.Port)

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
		log.Printf("Starting Chord node server on %s", node.Address)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	if flag_first_node {
		node.Successors = []string{node.Address}
	} else {
		// node.Successors = []string{nprime} TODO not sure about this one
	}

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

func validateInput(cfg Config) {
	flag_first_node = true

	if cfg.IPAddr == "" || net.ParseIP(cfg.IPAddr).To4() == nil {
		log.Fatal("-a must be a valid IPv4 address, e.g. 128.8.126.63")
	}
	if cfg.Port < 1024 || cfg.Port > 65535 {
		log.Fatal("-p must be in range 1024–65535")
	}

	if (cfg.JoinAddr == "" && cfg.JoinPort != 0) || (cfg.JoinAddr != "" && cfg.JoinPort == 0) {
		log.Fatal("--ja and --jp must be given together, or neither")
	}

	if cfg.JoinAddr != "" && cfg.JoinPort != 0 {
		flag_first_node = false
		if net.ParseIP(cfg.JoinAddr).To4() == nil {
			log.Fatal("--ja must be a valid IPv4 address")
		}
		if cfg.JoinPort < 1024 || cfg.JoinPort > 65535 {
			log.Fatal("--jp must be in range 1024–65535")
		}
	}

	if cfg.TS <= 1 || cfg.TS >= 60000 {
		log.Fatal("Ts must be specified, or between 1 and 60k")

	}
	if cfg.TFF <= 1 || cfg.TFF >= 60000 {
		log.Fatal("Tff must be specified, or between 1 and 60k")

	}
	if cfg.TCP <= 1 || cfg.TCP >= 60000 {
		log.Fatal("tcp must be specified, or between 1 and 60k")

	}
	if cfg.R <= 1 || cfg.R >= 32 {
		log.Fatal("R must be specified, or between 1 and 32")
	}
	if cfg.I != "" {
		if len(cfg.I) != 40 || !isHex(cfg.I) {
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
