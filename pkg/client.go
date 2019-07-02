package redis

import "net"

type client struct {
	id   clientId
	conn *conn
}

type clientId uint32

// newClient creates a new client from a net.Conn and returns the pointer of it.
func (r *Redis) newClient(conn net.Conn) *client {
	return &client{
		id:   r.stat.nextClientId(), // generate id
		conn: newConn(conn),         // connection
	}
}
