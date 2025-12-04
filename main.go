package main

import (
	"flag"
	"log"
	"net"
	"strings"
)

var (
	IP_addr      string
	port         int
	joinAddr     string
	joinPort     int
	ts           int
	tff          int
	tcp          int
	r            int
	i            string
	localaddress string // TODO to be removed
)

func init() {

}

func main() {
	flag.StringVar(&IP_addr, "a", "127.0.0.1", "Chord IP Address")
	flag.IntVar(&port, "p", 1234, "The port that the Chord client will bind to and listen on. Represented as a base-10 integer.")

	flag.StringVar(&joinAddr, "ja", "", "The IP address of the machine running a Chord node. The Chord client will join this node’s ring")

	flag.IntVar(&joinPort, "jp", 0, "The port that an existing Chord node is bound to and listening on.")
	flag.IntVar(&ts, "ts", 3000, "The time in milliseconds between invocations of ‘stabilize’. Represented as a base-10 integer.")
	flag.IntVar(&tff, "tff", 1000, "The time in milliseconds between invocations of ‘fix fingers’.")
	flag.IntVar(&tcp, "tcp", 3000, "The time in milliseconds between invocations of check predecessor’")
	flag.IntVar(&r, "r", 4, "The number of successors maintained by the Chord client. ")
	flag.StringVar(&i, "i", "", "he identifier (ID) assigned to the Chord client which will override the ID computed by the SHA1 sum of the client’s IP address and port number.")

	flag.Parse()

	validatInput()

	go RunShell(nil)

}
func RunShell(node *Node) {}










func validatInput() {

	if IP_addr == "" || net.ParseIP(IP_addr).To4() == nil {
		log.Fatal("-a must be a valid IPv4 address, e.g. 128.8.126.63")
	}
	if port < 1024 || port > 65535 {
		log.Fatal("-p must be in range 1024–65535")
	}

	if (joinAddr == "" && joinPort != 0) || (joinAddr != "" && joinPort == 0) {
		log.Fatal("--ja and --jp must be given together, or neither")
	}

	// 2) if both given, validate them
	if joinAddr != "" && joinPort != 0 {
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
