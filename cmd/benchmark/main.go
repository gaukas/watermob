package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gaukas/watermob"
)

const (
	defaultNetwork = "tcp"
	defaultRaddr   = ""
)

var (
	network  string
	raddr    string
	wasmPath string

	messageSize  int
	totalMessage int
	interval     time.Duration
)

func init() {
	flag.StringVar(&network, "network", defaultNetwork, "network type (tcp, udp, etc)")
	flag.StringVar(&network, "net", defaultNetwork, "network type (tcp, udp, etc) (shorthand)")
	flag.StringVar(&raddr, "raddr", "", "remote address to dial")
	flag.StringVar(&raddr, "a", "", "remote address to dial (shorthand)")
	flag.StringVar(&wasmPath, "webassembly-path", "", "path to the wasm file")
	flag.StringVar(&wasmPath, "wasm", "", "path to the wasm file (shorthand)")

	flag.IntVar(&messageSize, "message-size", 1024, "size of the message to send")
	flag.IntVar(&messageSize, "sz", 1024, "size of the message to send (shorthand)")
	flag.IntVar(&totalMessage, "total-message", 1000, "total number of messages to send")
	flag.IntVar(&totalMessage, "m", 1000, "total number of messages to send (shorthand)")
	flag.DurationVar(&interval, "interval", 1*time.Millisecond, "minimal interval between each message, ignored for commands other than echo")
	flag.DurationVar(&interval, "i", 1*time.Millisecond, "minimal interval between each message, ignored for commands other than echo (shorthand)")
}

func exitWithUsage() {
	flag.Usage()
	fmt.Println("To run the benchmark: benchmark type command [arguments...]")
	fmt.Printf("Possible types: pressure, echo\n")
	fmt.Printf("Possible commands: write, read\n")
	os.Exit(1)
}

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
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

	bd := watermob.NewBenchmarkDialer().SetMessageSize(messageSize).SetTotalMessage(totalMessage).SetInterval(interval)

	var writeBench bool
	benchCommand := flag.Arg(1)
	switch benchCommand {
	case "write":
		writeBench = true
	case "read":
		writeBench = false
	default:
		exitWithUsage()
	}

	benchType := flag.Arg(0)
	switch benchType {
	case "pressure":
		bd.PressureBenchmarkWATER(network, raddr, wasm, writeBench)
	case "echo":
		bd.EchoBenchmarkWATER(network, raddr, wasm, writeBench)
	default:
		exitWithUsage()
	}
}
