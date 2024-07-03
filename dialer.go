package watermob

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"syscall"
	"time"

	"github.com/refraction-networking/water"
	_ "github.com/refraction-networking/water/transport/v0"
	_ "github.com/refraction-networking/water/transport/v1"
	"github.com/tetratelabs/wazero"
)

var ErrNoDialer = errors.New("no dialer available")

var compilationCache = wazero.NewCompilationCache()

type Dialer struct {
	dial func(network, address string) (net.Conn, error) // if protector is set, this will be a protected dial function
	// protectedDial   func(network, address string) (net.Conn, error)
	// unprotectedDial func(network, address string) (net.Conn, error)

	configJSON []byte
	configPB   []byte

	forceInterpreter bool
}

func NewDialer() *Dialer {
	water.SetGlobalCompilationCache(compilationCache)

	return &Dialer{
		dial: net.Dial,
		// protectedDial: func(network, address string) (net.Conn, error) {
		// 	return nil, ErrNoDialer
		// },
		// unprotectedDial: net.Dial,
	}
}

// SetProtector updates the protectedDial function to use the provided Protector
// to protect the file descriptor of the connection.
func (d *Dialer) SetProtector(p Protector) {
	d.dial = func(network, address string) (net.Conn, error) {
		dialer := &net.Dialer{
			Timeout:   time.Second * 16,
			LocalAddr: nil,
			KeepAlive: 0,
			Control: func(network, address string, c syscall.RawConn) error {
				var innerErr error
				if err := c.Control(func(fd uintptr) {
					ok := p.Protect(int(fd))
					if !ok {
						innerErr = errors.New("failed to protect fd")
					}
				}); err != nil {
					return err
				}
				return innerErr
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

func (d *Dialer) DialWATER(network, remoteAddr string, wasm []byte) (NetConn, error) {
	config := &water.Config{
		TransportModuleBin: wasm,
		NetworkDialerFunc:  d.dial,
	}

	if len(config.TransportModuleBin) == 0 {
		return nil, errors.New("water: WebAssembly Transport Module binary is not provided in config")
	}

	if d.configJSON != nil {
		if err := config.UnmarshalJSON(d.configJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	} else if d.configPB != nil {
		if err := config.UnmarshalProto(d.configPB); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	if runtime.GOOS == "ios" || d.forceInterpreter {
		// Force-enable interpreter mode on iOS until we have a better workaround.
		config.RuntimeConfig().Interpreter()
	}

	ctx := context.Background()

	dialer, err := water.NewDialerWithContext(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dialer: %w", err)
	}

	conn, err := dialer.DialContext(ctx, network, remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	// conn is a net.Conn that you are familiar with.
	// So effectively, W.A.T.E.R. API ends here and everything below
	// this line is just how you treat a net.Conn.

	return &netConn{conn}, nil
}

func (d *Dialer) ForceInterpreter() {
	d.forceInterpreter = true
}

func (d *Dialer) DoNotForceInterpreter() {
	d.forceInterpreter = false
}
