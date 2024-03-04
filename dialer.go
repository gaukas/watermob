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

func unprotectedDial(network, address string) (net.Conn, error) {
	return net.Dial(network, address)
}

var protectedDial = func(network, address string) (net.Conn, error) {
	return nil, ErrNoDialer
}

type Protector interface {
	Protect(fd int) bool
}

// SetProtector updates the protectedDial function to use the provided Protector
// to protect the file descriptor of the connection.
func SetProtector(p Protector) {
	protectedDial = func(network, address string) (net.Conn, error) {
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

func ProtectedDialWATER(remoteAddr string, wasm []byte) (net.Conn, error) {
	return dialWATER(remoteAddr, wasm, protectedDial)
}

func UnprotectedDialWATER(remoteAddr string, wasm []byte) (net.Conn, error) {
	return dialWATER(remoteAddr, wasm, unprotectedDial)
}

func dialWATER(remoteAddr string, wasm []byte, dialerFunc func(network string, address string) (net.Conn, error)) (net.Conn, error) {
	config := &water.Config{
		TransportModuleBin: wasm,
		NetworkDialerFunc:  dialerFunc,
	}
	// configuring the standard out of the WebAssembly instance to inherit
	// from the parent process
	config.ModuleConfig().InheritStdout()
	config.ModuleConfig().InheritStderr()

	ctx := context.Background()
	// // optional: enable wazero logging
	// ctx = context.WithValue(ctx, experimental.FunctionListenerFactoryKey{},
	// 	logging.NewHostLoggingListenerFactory(os.Stderr, logging.LogScopeFilesystem|logging.LogScopePoll|logging.LogScopeSock))

	dialer, err := water.NewDialerWithContext(ctx, config)
	if err != nil {
		panic(fmt.Sprintf("failed to create dialer: %v", err))
	}

	conn, err := dialer.DialContext(ctx, "tcp", remoteAddr)
	if err != nil {
		panic(fmt.Sprintf("failed to dial: %v", err))
	}
	defer conn.Close()
	// conn is a net.Conn that you are familiar with.
	// So effectively, W.A.T.E.R. API ends here and everything below
	// this line is just how you treat a net.Conn.

	return conn, nil
}

func StartWorker(conn net.Conn) {
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
				return // connection closed
			}
			log.Printf("peer: %x\n", msg)
		case <-ticker.C:
			n, err := rand.Read(sendBuf)
			if err != nil {
				log.Printf("rand.Read: error %v, tearing down connection...", err)
				return
			}
			// print the bytes sending as hex string
			log.Printf("sending: %x\n", sendBuf[:n])

			_, err = conn.Write(sendBuf[:n])
			if err != nil {
				log.Printf("write: error %v, tearing down connection...", err)
				return
			}
		}
	}
}
