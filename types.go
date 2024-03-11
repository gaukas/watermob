package watermob

import (
	"net"
	"time"
)

// This file contains the explicitly defined types for the watermob package that
// maps to standard library types. This is done to allow for easy swapping of

// net.Addr
type NetAddr interface {
	Network() string
	String() string
}

// net.Conn
type NetConn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalAddr() NetAddr
	RemoteAddr() NetAddr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type netConn struct {
	embeddedConn net.Conn
}

func (c *netConn) Read(b []byte) (n int, err error) {
	return c.embeddedConn.Read(b)
}

func (c *netConn) Write(b []byte) (n int, err error) {
	return c.embeddedConn.Write(b)
}

func (c *netConn) Close() error {
	return c.embeddedConn.Close()
}

func (c *netConn) LocalAddr() NetAddr {
	return c.embeddedConn.LocalAddr()
}

func (c *netConn) RemoteAddr() NetAddr {
	return c.embeddedConn.RemoteAddr()
}

func (c *netConn) SetDeadline(t time.Time) error {
	return c.embeddedConn.SetDeadline(t)
}

func (c *netConn) SetReadDeadline(t time.Time) error {
	return c.embeddedConn.SetReadDeadline(t)
}

func (c *netConn) SetWriteDeadline(t time.Time) error {
	return c.embeddedConn.SetWriteDeadline(t)
}

func NewNetConn(c net.Conn) NetConn {
	return &netConn{c}
}
