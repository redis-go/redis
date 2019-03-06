package redis

import (
	"fmt"
	"net"
)

// Redis represents a redis instance.
type Redis struct {
	s  *server
	cc *clientController
}

// NewRedis creates and returns a new Redis instance.
func NewRedis(options ...Option) *Redis {
	r := &Redis{cc: newClientController()}
	r.s = newServer(r.handleConn)

	// apply options
	for _, o := range options {
		o(r)
	}

	// defaults
	if r.s.addr == "" {
		WithAddr(fmt.Sprintf("localhost:%d", DefaultPort))(r)
	}

	return r
}

// Run creates a new Redis and runs the server.
func Run(options ...Option) error {
	return NewRedis(options...).Run()
}

// Run runs the Redis server.
func (r *Redis) Run() error {
	return r.s.listenAndServe()
}

// Shutdown shuts down the Redis.
func (r *Redis) Shutdown() error {
	return r.s.close()
}

// Option is a option to configure a new Redis.
type Option func(r *Redis)

// handleConn handles incoming net.Conn connections.
func (r *Redis) handleConn(conn net.Conn) {
	// Create client.
	client := newClient(newConn(conn))
	r.cc.addClient(client)

	// Read
	// Exec Commands
	// Write
}
