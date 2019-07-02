package redis

import (
	"fmt"
	"go.uber.org/zap"
	"net"
)

// Redis represents a redis instance.
type Redis struct {
	server  *server
	cmds    *cmdProcessor
	clients *clientController
	stat    *stats
	logger  *zap.Logger
}

// NewRedis creates and returns a new Redis instance.
func NewRedis(options ...Option) *Redis {
	r := &Redis{}
	r.logger, _ = zap.NewProductionConfig().Build()
	r.clients = newClientController()
	r.server = newServer(r.handleConn, r.logger)
	r.cmds = newCmdProcessor(r.logger)

	// apply options
	for _, o := range options {
		o(r)
	}

	// defaults
	if r.server.addr == "" {
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
	r.logger.Info("Listening for new clients...")
	return r.server.listenAndServe()
}

// Shutdown shuts down the Redis.
func (r *Redis) Shutdown() error {
	return r.server.close()
}

// Logger returns the Redis logger
func (r *Redis) Logger() *zap.Logger {
	return r.logger
}

// Option is a option to configure a new Redis.
type Option func(r *Redis)

// handleConn handles incoming net.Conn connections.
func (r *Redis) handleConn(conn net.Conn) {
	// Create client.
	client := r.newClient(conn)
	// Add client to clientController
	r.clients.addClient(client)

	// Remove client when all done.
	defer func() {
		/*if err != errDetached {
			// do not close the connection when a detach is detected.
			c.conn.Close()
		}*/
		client.conn.Close()
		r.clients.removeClient(client.id)
	}()

	// Read and execute commands from client connection.
	for cmd := range client.conn.rd.ReadCommands() { // cmds are send through the channel as they are executed
		args := cmd.Args() // args, has command name at index 0
		var params [][]byte
		if len(args) == 1 { // command got no parameters passed
			params = make([][]byte, 0)
		} else {
			params = args[1:]
		}
		r.cmds.run(string(args[0]), client, params)
	}
}
