package watermob

import (
	"log"
	"time"

	"github.com/gaukas/benchmarkconn"
)

type BenchmarkDialer struct {
	*Dialer

	messageSize   int
	totalMessage  int
	intervalMicro int
}

func NewBenchmarkDialer() *BenchmarkDialer {
	return &BenchmarkDialer{
		Dialer: NewDialer(),
	}
}

func (d *BenchmarkDialer) SetMessageSize(size int) *BenchmarkDialer {
	d.messageSize = size
	return d
}

func (d *BenchmarkDialer) SetTotalMessage(total int) *BenchmarkDialer {
	d.totalMessage = total
	return d
}

func (d *BenchmarkDialer) SetInterval(interval time.Duration) *BenchmarkDialer {
	d.intervalMicro = int(interval.Microseconds())
	return d
}

func (d *BenchmarkDialer) BenchmarkWATERWrite(network, remoteAddr string, wasm []byte) error {
	conn, err := d.DialWATER(network, remoteAddr, wasm)
	if err != nil {
		return err
	}
	defer conn.Close()

	benchmark := &benchmarkconn.PressuredBenchmark{
		MessageSize:   d.messageSize,
		TotalMessages: uint64(d.totalMessage),
	}

	if err := benchmark.Writer(conn.(*netConn).embeddedConn); err != nil {
		return err
	}

	log.Printf("BenchmarkWATERWrite Result: %v", benchmark.Result())
	return nil
}

func (d *BenchmarkDialer) BenchmarkWATERRead(network, remoteAddr string, wasm []byte) error {
	conn, err := d.DialWATER(network, remoteAddr, wasm)
	if err != nil {
		return err
	}
	defer conn.Close()

	benchmark := &benchmarkconn.PressuredBenchmark{
		MessageSize:   d.messageSize,
		TotalMessages: uint64(d.totalMessage),
	}

	if err := benchmark.Reader(conn.(*netConn).embeddedConn); err != nil {
		return err
	}

	log.Printf("BenchmarkWATERRead Result: %v", benchmark.Result())
	return nil
}

func (d *BenchmarkDialer) BenchmarkWATEREcho(network, remoteAddr string, wasm []byte) error {
	conn, err := d.DialWATER(network, remoteAddr, wasm)
	if err != nil {
		return err
	}
	defer conn.Close()

	benchmark := &benchmarkconn.IntervalBenchmark{
		MessageSize:   d.messageSize,
		TotalMessages: uint64(d.totalMessage),
		Interval:      time.Duration(d.intervalMicro) * time.Microsecond,
		Echo:          true,
	}

	if err := benchmark.Writer(conn.(*netConn).embeddedConn); err != nil {
		return err
	}

	log.Printf("BenchmarkWATEREcho Result: %v", benchmark.Result())
	return nil
}
