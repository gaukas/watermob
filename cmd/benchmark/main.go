package main

import (
	"fmt"
	"os"
)

const (
	defaultNetwork = "tcp"
	defaultRaddr   = ""
)

func Usage() {
	fmt.Println("Example: benchmark <type> <operation> <remote_addr> [arguments...]")
	fmt.Printf("Possible <type>: pressure, echo\n")
	fmt.Printf("Possible <operation>: write, read\n")
}

func main() {
	args := os.Args[1:]

	if len(args) < 3 {
		Usage()
		os.Exit(1)
	}

	b := NewBenchmark()

	benchType := os.Args[1]
	benchOp := os.Args[2]
	remoteAddr := os.Args[3]

	b.SetBenchType(benchType)
	b.SetCommand(benchOp)
	b.SetRemoteAddress(remoteAddr)
	if err := b.Init(os.Args[4:]); err != nil {
		fmt.Printf("Failed to initialize benchmark: %v\n", err)
		os.Exit(1)
	}

	if err := b.Run(); err != nil {
		fmt.Printf("Failed to run benchmark: %v\n", err)
		os.Exit(1)
	}
}
