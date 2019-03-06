package redis

import "net"

// conn represents a client connection.
type conn struct {
	conn net.Conn
	w    Writer
	// r        reader
	detached bool
	closed   bool
	// cmds     []Command
}

func (c *conn) WriteNull() {
	panic("implement me")
}

// newConn returns a new Conn.
func newConn(c net.Conn) *conn {
	return &conn{
		conn: c,
	}
}

// NetConn returns the base net.Conn.
func (c *conn) NetConn() net.Conn {
	return c.conn
}

// Close flushes the buffer and closes the connection.
func (c *conn) Close() error {
	// c.w.Flush()
	c.closed = true
	return c.conn.Close()
}
