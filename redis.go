package redis

import (
	"crypto/tls"
	"fmt"
	"github.com/redis-go/redcon"
	"github.com/redis-go/redis/cmds"
	"github.com/redis-go/redis/store"
	"strings"
	"sync"
	"time"
)

const (
	SyntaxERR     = "ERR syntax error"
	InvalidIntErr = "ERR value is not an integer or out of range"
	WrongTypeErr  = "WRONGTYPE Operation against a key holding the wrong kind of value"
)

// This is the redis server.
type Redis struct {
	// used for normal keys
	redisDb *ItemStore

	// Locking is important, share this mutex around to provide state.
	mu *sync.RWMutex

	commands       Commands
	unknownCommand UnknownCommand

	handler Handler

	accept  Accept
	onClose OnClose

	// TODO version
	// TODO log writer
	// TODO modules
}

// A Handler is called when a request is received and after Accept
// (if Accept allowed the connection by returning true).
//
// For implementing an own handler see the default handler
// as a perfect example in the createDefault() function.
type Handler func(c redcon.Conn, cmd redcon.Command, r *Redis)

// Accept is called when a client tries to connect and before everything else,
// the client connection will be closed instantaneously if the function returns false.
type Accept func(c redcon.Conn, r *Redis) bool

// OnClose is called when a client connection is closed.
type OnClose func(c redcon.Conn, err error, r *Redis)

// Commands map
type Commands map[string]CommandHandler

// The CommandHandler is triggered when the received
// command equals a registered command.
//
// However the CommandHandler is executed by the Handler,
// so if you implement an own Handler make sure the CommandHandler is called.
type CommandHandler func(c redcon.Conn, cmd redcon.Command, r *Redis)

// Is called when a request is received,
// after Accept and if the command is not registered.
//
// However UnknownCommand is executed by the Handler,
// so if you implement an own Handler make sure to include UnknownCommand.
type UnknownCommand func(c redcon.Conn, cmd redcon.Command, r *Redis)

// Item stores should implement this interface.
type ItemStore interface {
	// Get the redis instance.
	Redis() *Redis

	// Sets a key with an item.
	Set(key *string, i *Item)
	// Returns the item by the key or nil if key does not exists.
	Get(key *string) *Item
	// Deletes a key, returns true if key existed.
	Delete(key *string) bool
	// Check if key exists.
	Exists(key *string) bool

	// Check if the key is expired and deletes the key if so.
	// Returns true if the key did existed and is expired.
	CheckExpire(key *string) bool
}

// An item type should implement this interface.
type Item interface {
	// The pointer to the value.
	Value() interface{}
	// The type of the Item.
	ValueType() string
	// Get timestamp when the item expires.
	Expire() time.Time
	// Check if this item is expired.
	Expired() bool

	// OnDelete is triggered before the key of the item is deleted.
	// Returning an error does not cancel the deletion of the key by default and should never.
	OnDelete(key *string, r *Redis) error
}

// Run runs the default redis server.
// Initializes the default redis if not already.
func Run(addr string) error {
	return Default().Run(addr)
}

// Run runs the redis server.
func (r *Redis) Run(addr string) error {
	return redcon.ListenAndServe(
		addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			r.HandlerFn()(conn, cmd, r)
		},
		func(conn redcon.Conn) bool {
			return r.AcceptFn()(conn, r)
		},
		func(conn redcon.Conn, err error) {
			r.OnCloseFn()(conn, err, r)
		},
	)
}

// Run runs the redis server with tls.
func (r *Redis) RunTLS(addr string, tls *tls.Config) error {
	return redcon.ListenAndServeTLS(
		addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			r.HandlerFn()(conn, cmd, r)
		},
		func(conn redcon.Conn) bool {
			return r.AcceptFn()(conn, r)
		},
		func(conn redcon.Conn, err error) {
			r.OnCloseFn()(conn, err, r)
		},
		tls,
	)
}

// Gets the handler func.
func (r *Redis) HandlerFn() Handler {
	r.Mu().RLock()
	defer r.Mu().RUnlock()
	return r.handler
}

// Sets the handler func.
// Live updates (while redis is running) works.
func (r *Redis) SetHandlerFn(new Handler) {
	r.Mu().Lock()
	defer r.Mu().Unlock()
	r.handler = new
}

// Gets the accept func.
func (r *Redis) AcceptFn() Accept {
	r.Mu().RLock()
	defer r.Mu().RUnlock()
	return r.accept
}

// Sets the accept func.
// Live updates (while redis is running) works.
func (r *Redis) SetAcceptFn(new Accept) {
	r.Mu().Lock()
	defer r.Mu().Unlock()
	r.accept = new
}

// Gets the onclose func.
func (r *Redis) OnCloseFn() OnClose {
	r.Mu().RLock()
	defer r.Mu().RUnlock()
	return r.onClose
}

// Sets the onclose func.
// Live updates (while redis is running) works.
func (r *Redis) SetOnCloseFn(new OnClose) {
	r.Mu().Lock()
	defer r.Mu().Unlock()
	r.onClose = new
}

// The mutex of the redis.
func (r *Redis) Mu() *sync.RWMutex {
	return r.mu
}

// RegisterCommand adds a command to the redis instance.
// If cmd exists already the handler is overridden.
func (r *Redis) RegisterCommand(cmd string, handler CommandHandler) {
	r.Mu().Lock()
	defer r.Mu().Unlock()
	r.getCommands()[cmd] = handler
}

// UnregisterCommand removes a command.
func (r *Redis) UnregisterCommand(cmd string) {
	r.Mu().Lock()
	defer r.Mu().Unlock()
	delete(r.commands, cmd)
}

// Commands returns the commands map.
func (r *Redis) Commands() Commands {
	r.Mu().RLock()
	defer r.Mu().RUnlock()
	return r.getCommands()
}
func (r *Redis) getCommands() Commands {
	return r.commands
}

// CommandExists checks if one or more commands are registered.
func (r *Redis) CommandExists(cmds ...string) bool {
	// does this make the performance better because it does not create a loop every time?
	if len(cmds) == 1 {
		_, ex := r.Commands()[cmds[0]]
		return ex
	}

	for _, cmd := range cmds {
		if _, ex := r.Commands()[cmd]; !ex {
			return false
		}
	}
	return true
}

// GetCommandHandler returns the CommandHandler of cmd.
func (r *Redis) GetCommandHandler(cmd string) CommandHandler {
	r.Mu().RLock()
	defer r.Mu().RUnlock()
	return r.getCommands()[cmd]
}

// FlushCommands removes all commands.
func (r *Redis) FlushCommands() {
	r.Mu().Lock()
	defer r.Mu().Unlock()
	r.commands = make(Commands)
}

var defaultRedis *Redis

// Default redis server.
// Initializes the default redis if not already.
// You can change the fields or value behind the pointer
// of the returned redis pointer to extend/change the default.
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
	cmnds := Commands{
		"ping": cmds.Ping,
		"set":  cmds.Set,
		"get":  cmds.Get,
		"del":  cmds.Del,
		"ttl":  cmds.Ttl,
	}
	return &Redis{
		mu: new(sync.RWMutex),
		accept: func(c redcon.Conn, r *Redis) bool {
			return true
		},
		onClose: func(c redcon.Conn, err error, r *Redis) {
		},
		handler: func(c redcon.Conn, cmd redcon.Command, r *Redis) {
			P("-------------------------")
			P(string(cmd.Raw))
			cmdl := strings.ToLower(string(cmd.Args[0]))
			if r.CommandExists(cmdl) {
				r.GetCommandHandler(cmdl)(c, cmd, r)
			} else {
				r.unknownCommand(c, cmd, r)
			}
		},
		unknownCommand: func(c redcon.Conn, cmd redcon.Command, r *Redis) {
			c.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
		},
		internalItemStore: store.NewDevStore(),
		externalItemStore: store.NewDevStore(),
		commands:          cmnds,
	}
}

func P(s string) {
	fmt.Println(s)
}
