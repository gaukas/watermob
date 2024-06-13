package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	. "github.com/gaukas/watermob"
)

const (
	defaultNetwork = "tcp"
	defaultRaddr   = ""
)

var (
	network  string
	raddr    string
	wasmPath string
	command  string

	messageSize  int
	totalMessage int
	interval     time.Duration
)

func init() {
	flag.StringVar(&network, "network", defaultNetwork, "network type (tcp, udp, etc)")
	flag.StringVar(&network, "n", defaultNetwork, "network type (tcp, udp, etc) (shorthand)")
	flag.StringVar(&raddr, "raddr", "", "remote address to dial")
	flag.StringVar(&raddr, "a", "", "remote address to dial (shorthand)")
	flag.StringVar(&wasmPath, "wasm", "", "path to the wasm file")
	flag.StringVar(&wasmPath, "w", "", "path to the wasm file (shorthand)")

	flag.IntVar(&messageSize, "message-size", 1024, "size of the message to send")
	flag.IntVar(&messageSize, "s", 1024, "size of the message to send (shorthand)")
	flag.IntVar(&totalMessage, "total-message", 1000, "total number of messages to send")
	flag.IntVar(&totalMessage, "t", 1000, "total number of messages to send (shorthand)")
	flag.DurationVar(&interval, "interval", 1*time.Millisecond, "minimal interval between each message, ignored for commands other than echo")
	flag.DurationVar(&interval, "i", 1*time.Millisecond, "minimal interval between each message, ignored for commands other than echo (shorthand)")
}

func exitWithUsage() {
	flag.Usage()
	fmt.Println("To run the benchmark: benchmark command [arguments...]")
	fmt.Printf("Possible commands: write, read, echo\n")
	os.Exit(1)
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		exitWithUsage()
	}

	if wasmPath == "" {
		fmt.Println("wasm file path is required")
		exitWithUsage()
	}
	wasm, err := os.ReadFile(wasmPath)
	if err != nil {
		panic(fmt.Sprintf("failed to read wasm file: %v", err))
	}

	if raddr == "" {
		fmt.Println("remote address is required")
		exitWithUsage()
	}

	bd := NewBenchmarkDialer().SetMessageSize(messageSize).SetTotalMessage(totalMessage).SetInterval(interval)

	command = flag.Arg(0)
	switch command {
	case "write":
		bd.BenchmarkWATERWrite(network, raddr, wasm)
	case "read":
		bd.BenchmarkWATERRead(network, raddr, wasm)
	case "echo":
		bd.BenchmarkWATEREcho(network, raddr, wasm)
	default:
		exitWithUsage()
	}
}
