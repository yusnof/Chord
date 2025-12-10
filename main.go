package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
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

	Id := hash(FormatToString(cfg.IPAddr, cfg.Port))

	if cfg.I != "" {
		n := new(big.Int)
		n, ok := n.SetString(cfg.I, 10)
		if !ok {
			log.Fatal("error in transforming values")
		}
		Id = n
	}
	node := &Node{
		IP:   cfg.IPAddr,
		Port: cfg.Port,
		ID:   Id,
	}
	//TOD add more logic here
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Port))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	rpc.Register(node)
	rpc.HandleHTTP()

	go func() {
		log.Printf("Starting Chord node server on %s", strconv.Itoa(cfg.Port))
		if err := http.Serve(lis, nil); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	if cfg.Flag_first_node {
		node.Create()
	} else {

		JoinAddr := Node{
			IP:   cfg.JoinAddr,
			Port: cfg.JoinPort,
			ID:   hash(FormatToString(cfg.JoinAddr, cfg.JoinPort)),
		}

		node.Join(&JoinAddr)
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
