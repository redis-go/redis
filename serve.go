package redis

import (
	"crypto/tls"
	"github.com/redis-go/redcon"
	"time"
)

// Run runs the default redis server.
// Initializes the default redis if not already.
func Run(addr string) error {
	return Default().Run(addr)
}

// Run runs the redis server.
func (r *Redis) Run(addr string) error {
	go r.KeyExpirer().Start(100*time.Millisecond, 20, 25)
	return redcon.ListenAndServe(
		addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			r.HandlerFn()(r.NewClient(conn), cmd)
		},
		func(conn redcon.Conn) bool {
			return r.AcceptFn()(r.NewClient(conn))
		},
		func(conn redcon.Conn, err error) {
			r.OnCloseFn()(r.NewClient(conn), err)
		},
	)
}

// Run runs the redis server with tls.
func (r *Redis) RunTLS(addr string, tls *tls.Config) error {
	go r.KeyExpirer().Start(100*time.Millisecond, 20, 25)
	return redcon.ListenAndServeTLS(
		addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			r.HandlerFn()(r.NewClient(conn), cmd)
		},
		func(conn redcon.Conn) bool {
			return r.AcceptFn()(r.NewClient(conn))
		},
		func(conn redcon.Conn, err error) {
			r.OnCloseFn()(r.NewClient(conn), err)
		},
		tls,
	)
}
