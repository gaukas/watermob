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

func (d *BenchmarkDialer) PressureBenchmarkWATER(network, remoteAddr string, wasm []byte, write bool) error {
	conn, err := d.DialWATER(network, remoteAddr, wasm)
	if err != nil {
		return err
	}
	defer conn.Close()

	benchmark := &benchmarkconn.PressuredBenchmark{
		MessageSize:   d.messageSize,
		TotalMessages: uint64(d.totalMessage),
	}

	if write {
		if err := benchmark.Writer(conn.(*netConn).embeddedConn); err != nil {
			return err
		}
		time.Sleep(10 * time.Second)
	} else {
		if err := benchmark.Reader(conn.(*netConn).embeddedConn); err != nil {
			return err
		}
	}

	log.Printf("PressureBenchmarkWATER Result: %v", benchmark.Result())
	return nil
}

func (d *BenchmarkDialer) EchoBenchmarkWATER(network, remoteAddr string, wasm []byte, write bool) error {
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

	if write {
		if err := benchmark.Writer(conn.(*netConn).embeddedConn); err != nil {
			return err
		}
		time.Sleep(10 * time.Second)
	} else {
		if err := benchmark.Reader(conn.(*netConn).embeddedConn); err != nil {
			return err
		}
	}

	log.Printf("EchoBenchmarkWATER Result: %v", benchmark.Result())
	return nil
}
