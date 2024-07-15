package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gaukas/watermob"
)

func NewBenchmark() *Benchmark {
	b := &Benchmark{
		fs: flag.NewFlagSet("", flag.ContinueOnError),
	}

	b.interpreter = b.fs.Bool("I", false, "force interpreter mode")
	b.network = b.fs.String("net", defaultNetwork, "network type (tcp, udp, etc)")
	b.wasmPath = b.fs.String("wasm", "", "path to the wasm file")
	b.messageSz = b.fs.Int("sz", 1024, "size of the message to send/expect")
	b.totalMsg = b.fs.Int("m", 1000, "total number of messages to send/expect")
	b.interval = b.fs.Duration("i", 1*time.Millisecond, "minimal interval between each message, only for echo")
	b.waterConf = b.fs.String("c", "", "path to the WATER config file")
	return b
}

type Benchmark struct {
	fs *flag.FlagSet

	interpreter *bool

	benchType string
	command   string
	raddr     string

	network  *string
	wasmPath *string

	messageSz *int
	totalMsg  *int

	wasm      []byte
	waterConf *string

	interval *time.Duration

	bd *watermob.BenchmarkDialer
}

func (b *Benchmark) Usage() {
	Usage()
	b.fs.Usage()
}

func (b *Benchmark) RemoteAddress() string {
	return b.raddr
}

func (b *Benchmark) BenchType() string {
	return b.benchType
}

func (b *Benchmark) Command() string {
	return b.command
}

func (b *Benchmark) SetBenchType(benchType string) {
	b.benchType = benchType
}

func (b *Benchmark) SetCommand(command string) {
	b.command = command
}

func (b *Benchmark) SetRemoteAddress(raddr string) {
	b.raddr = raddr
}

func (b *Benchmark) Init(args []string) error {
	if err := b.fs.Parse(args); err != nil {
		return err
	}

	if *b.wasmPath == "" {
		return errors.New("--wasm is required")
	}

	var err error
	b.wasm, err = os.ReadFile(*b.wasmPath)
	if err != nil {
		return fmt.Errorf("failed to read wasm file at %s (specified by --wasm): %v ", *b.wasmPath, err)
	}

	b.bd = watermob.NewBenchmarkDialer().SetMessageSize(*b.messageSz).SetTotalMessage(*b.totalMsg).SetInterval(*b.interval)

	if *b.waterConf != "" {
		watmConf, err := os.ReadFile(*b.waterConf)
		if err != nil {
			return fmt.Errorf("failed to read watm config file at %s (specified by --c): %v ", *b.waterConf, err)
		}
		b.bd.SetConfigJSON(watmConf)
	}

	if *b.interpreter {
		b.bd.ForceInterpreter()
	}

	return nil
}

func (b *Benchmark) Run() error {
	var writeBench bool
	switch b.command {
	case "write":
		writeBench = true
	case "read":
		writeBench = false
	default:
		b.Usage()
		return nil
	}

	switch b.benchType {
	case "pressure":
		if err := b.bd.PressureBenchmarkWATER(*b.network, b.raddr, b.wasm, writeBench); err != nil {
			return err
		}
	case "echo":
		if err := b.bd.EchoBenchmarkWATER(*b.network, b.raddr, b.wasm, writeBench); err != nil {
			return err
		}
	default:
		b.Usage()
		return nil
	}

	return nil
}
