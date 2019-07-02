package redis

import (
	"github.com/redis-go/redis/pkg/proto"
	"net"
)

// conn represents a client connection.
type conn struct {
	conn net.Conn
	//w    *writer
	rd       *proto.Reader
	detached bool
	closed   bool
}

// newConn returns a new Conn.
func newConn(c net.Conn) *conn {
	return &conn{
		conn: c,
		rd:   proto.NewReader(c),
	}
}

// NetConn returns the base net.Conn.
func (c *conn) NetConn() net.Conn {
	return c.conn
}

// Close closes the connection.
func (c *conn) Close() error {
	c.closed = true
	return c.conn.Close()
}
