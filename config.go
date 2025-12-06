package main
import (
    "flag"
    "log"
    "net"
	"os"
)
type Config struct {
    IPAddr   string
    Port     int
    JoinAddr string
    JoinPort int
    TS       int
    TFF      int
    TCP      int
    R        int
    I       string
    Flag_first_node bool
}

func LogerConfigurationSetup() *os.File {
	logFile, err := os.OpenFile("chord.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("logger is bad ")
	}

	log.SetOutput(logFile)

	// Add this to force immediate writes:
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return logFile
}

func Loadconfig() Config {
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

	validated_cfg := validateInput(&cfg)

	return *validated_cfg

}
func validateInput(cfg *Config) (*Config){
	cfg.Flag_first_node = true

	
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
		cfg.Flag_first_node= false
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
	return cfg 
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
