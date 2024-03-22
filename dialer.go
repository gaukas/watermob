package watermob

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net"
	"syscall"
	"time"

	"github.com/refraction-networking/water"
	_ "github.com/refraction-networking/water/transport/v0"
)

var ErrNoDialer = errors.New("no dialer available")

type Dialer struct {
	protectedDial   func(network, address string) (net.Conn, error)
	unprotectedDial func(network, address string) (net.Conn, error)

	configJSON []byte
	configPB   []byte
}

func NewDialer() *Dialer {
	return &Dialer{
		protectedDial: func(network, address string) (net.Conn, error) {
			return nil, ErrNoDialer
		},
		unprotectedDial: net.Dial,
	}
}

// SetProtector updates the protectedDial function to use the provided Protector
// to protect the file descriptor of the connection.
func (d *Dialer) SetProtector(p Protector) {
	d.protectedDial = func(network, address string) (net.Conn, error) {
		dialer := &net.Dialer{
			Timeout:   time.Second * 16,
			LocalAddr: nil,
			KeepAlive: 0,
			Control: func(network, address string, c syscall.RawConn) error {
				return c.Control(func(fd uintptr) {
					ok := p.Protect(int(fd))
					if !ok {
						panic("failed to protect fd")
					}
				})
			},
		}

		return dialer.Dial(network, address)
	}
}

func (d *Dialer) SetConfigJSON(configJSON []byte) {
	d.configJSON = configJSON
	d.configPB = nil
}

func (d *Dialer) SetConfigPB(configPB []byte) {
	d.configPB = configPB
	d.configJSON = nil
}

func (d *Dialer) DialWATERProtected(network, remoteAddr string, wasm []byte) (NetConn, error) {
	return d.dialWATER(network, remoteAddr, wasm, d.protectedDial)
}

func (d *Dialer) DialWATERUnprotected(network, remoteAddr string, wasm []byte) (NetConn, error) {
	return d.dialWATER(network, remoteAddr, wasm, d.unprotectedDial)
}

func (d *Dialer) DirectlyStartWorkerProtected(network, remoteAddr string, wasm []byte) error {
	conn, err := d.DialWATERProtected(network, remoteAddr, wasm)
	if err != nil {
		panic(fmt.Sprintf("failed to dial: %v", err))
	}

	return startWorker(conn)
}

func (d *Dialer) DirectlyStartWorkerUnprotected(network, remoteAddr string, wasm []byte) error {
	conn, err := d.DialWATERUnprotected(network, remoteAddr, wasm)
	if err != nil {
		panic(fmt.Sprintf("failed to dial: %v", err))
	}

	return startWorker(conn)
}

func (d *Dialer) dialWATER(network, remoteAddr string,
	wasm []byte,
	dialerFunc func(network, address string) (net.Conn, error),
) (NetConn, error) {
	config := &water.Config{
		TransportModuleBin: wasm,
		NetworkDialerFunc:  dialerFunc,
	}

	if len(config.TransportModuleBin) == 0 {
		return nil, errors.New("water: WebAssembly Transport Module binary is not provided in config")
	}

	if d.configJSON != nil {
		config.UnmarshalJSON(d.configJSON)
	} else if d.configPB != nil {
		config.UnmarshalProto(d.configPB)
	}

	ctx := context.Background()

	dialer, err := water.NewDialerWithContext(ctx, config)
	if err != nil {
		panic(fmt.Sprintf("failed to create dialer: %v", err))
	}

	conn, err := dialer.DialContext(ctx, network, remoteAddr)
	if err != nil {
		panic(fmt.Sprintf("failed to dial: %v", err))
	}
	defer conn.Close()
	// conn is a net.Conn that you are familiar with.
	// So effectively, W.A.T.E.R. API ends here and everything below
	// this line is just how you treat a net.Conn.

	return &netConn{conn}, nil
}

func startWorker(conn NetConn) error {
	defer conn.Close()

	log.Printf("Connected to %s", conn.RemoteAddr())
	chanMsgRecv := make(chan []byte, 4) // up to 4 messages in the buffer
	// start a goroutine to read data from the connection
	go func() {
		defer close(chanMsgRecv)
		buf := make([]byte, 1024) // 1 KiB
		for {
			n, err := conn.Read(buf)
			if err != nil {
				log.Printf("read conn: error %v, tearing down connection...", err)
				conn.Close()
				return
			}
			chanMsgRecv <- buf[:n]
		}
	}()

	// start a ticker for sending message every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var sendBuf []byte = make([]byte, 4) // 4 bytes per message
	for {
		select {
		case msg := <-chanMsgRecv:
			if msg == nil {
				return errors.New("connection closed")
			}
			log.Printf("peer: %x\n", msg)
		case <-ticker.C:
			n, err := rand.Read(sendBuf)
			if err != nil {
				log.Printf("rand.Read: error %v, tearing down connection...", err)
				return err
			}
			// print the bytes sending as hex string
			log.Printf("sending: %x\n", sendBuf[:n])

			_, err = conn.Write(sendBuf[:n])
			if err != nil {
				log.Printf("write: error %v, tearing down connection...", err)
				return err
			}
		}
	}
}
