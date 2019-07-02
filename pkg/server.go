package redis

import (
	"crypto/tls"
	"go.uber.org/zap"
	"net"
)

const (
	DefaultPort = 6379
)

// WithTLS option makes the Server use TLS.
func WithTLS(config *tls.Config) Option {
	return func(r *Redis) {
		r.server.tlsCfg = config
	}
}

// WithAddr option sets the address of the Server.
// This option is applied by default on localhost and redis standard port.
func WithAddr(addr string) Option {
	return func(r *Redis) {
		r.server.addr = addr
	}
}

// server listens for connections and passes them further to the handler.
type server struct {
	// Address the server listens on.
	addr string
	// The connection handler.
	handler connHandler
	// If non.nil the Server use tls on startup.
	tlsCfg *tls.Config
	// The current listener to accept new connections.
	ln net.Listener
	// The server logger.
	logger *zap.Logger
}

// ConnHandler is called in a new goroutine on every new connection.
type connHandler func(conn net.Conn)

// newServer returns a new server and when started runs handler in a new goroutine on any new connection.
func newServer(handler connHandler, logger *zap.Logger) *server {
	s := &server{
		handler: handler,
		logger:  logger.With(zap.String("sector", "server")),
	}
	return s
}

// listenAndServe starts listening and serving for new connections.
func (s *server) listenAndServe() error {
	var ln net.Listener
	var err error

	if s.tlsCfg == nil {
		s.logger.Sugar().Debugf("tcp listener on address %s", s.addr)
		ln, err = net.Listen("tcp", s.addr)
	} else {
		s.logger.Sugar().Debugf("tcp listener on address %s with tls config %+v", s.addr, s.tlsCfg)
		ln, err = tls.Listen("tcp", s.addr, s.tlsCfg)
	}

	if err != nil {
		return err
	}

	s.ln = ln
	return s.serve()
}

func (s *server) serve() error {
	s.logger.Debug("serving new connections...")
	for {
		c, err := s.ln.Accept()
		if err != nil {
			return err
		}
		go func() {
			s.logger.Debug("new connection", zap.String("addr", c.RemoteAddr().String()))
			s.handler(c)
		}()
	}
}

// close stops the Server from listening new connections.
func (s *server) close() error {
	return s.ln.Close()
}
