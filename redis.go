package redis

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/redis-go/redcon"
	"strings"
)

// This is the redis server.
type Redis struct {
	// this is called on every request to get the commands,
	// e.g. registering commands by loading redis modules while the server is running is supported
	Commands Commands

	// this is called when a request is received
	Handler func(conn redcon.Conn, cmd redcon.Command)
	// this is called when a client tries to connect,
	// the client connection will be closed instantaneously if the function returns false
	Accept func(conn redcon.Conn) bool
	// this is called when a client connection has been closed
	OnClose func(conn redcon.Conn, err error)
}

var defaultRedis *Redis

// Default redis server.
// You can alter the fields of the returned Redis struct to extend the default.
func Default() *Redis {
	if defaultRedis != nil {
		return defaultRedis
	}
	defaultRedis = createDefault()
	return defaultRedis
}

// createDefault creates a new default redis.
func createDefault() *Redis {
	// initialize default redis server
	r := new(Redis)

	// ping
	r.AddCommand("ping", func(c redcon.Conn, cmd redcon.Command) {
		if len(cmd.Args) > 1 {
			var buf bytes.Buffer
			for i := 1; i-1 < len(cmd.Args); i++ {
				buf.Write(cmd.Args[i])
				buf.WriteString(" ")
				fmt.Println(i)
			}
			s := buf.String()
			s = s[:len(s)-1]
			c.WriteString(s)
			return
		}
		c.WriteString("PONG")
	})
	r.AddCommand("set", func(c redcon.Conn, cmd redcon.Command) {
		c.WriteError("Not impl")
	})

	r.Handler = func(conn redcon.Conn, cmd redcon.Command) {
		cmdl := strings.ToLower(string(cmd.Args[0]))
		if r.CommandExists(cmdl) {
			r.GetCommandHandler(cmdl)(conn, cmd)
		}
	}
	r.Accept = func(conn redcon.Conn) bool {
		return true
	}
	return r
}

// Run runs the default redis server.
func Run(addr string) error {
	return Default().Run(addr)
}

// Run runs the redis server.
func (r *Redis) Run(addr string) error {
	return redcon.ListenAndServe(addr, r.Handler, r.Accept, r.OnClose)
}

// Run runs the redis server.
func (r *Redis) RunTLS(addr string, tsl *tls.Config) error {
	return redcon.ListenAndServeTLS(addr, r.Handler, r.Accept, r.OnClose, tsl)
}
