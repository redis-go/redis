package redis

import "github.com/tidwall/redcon"

type CommandHandler func(conn redcon.Conn, cmd redcon.Command)
type Commands map[string]CommandHandler

// AddCommand adds a command.
// If cmd exists already the handler is overridden.
func (r *Redis) AddCommand(cmd string, handler CommandHandler) {
	r.GetCommands()[cmd] = handler
}

// RemoveCommand removes a command.
func (r *Redis) RemoveCommand(cmd string) {
	delete(r.Commands, cmd)
}

// GetCommands returns the command map.
func (r *Redis) GetCommands() Commands {
	if r.Commands == nil {
		r.Commands = make(Commands)
	}
	return r.Commands
}

// CommandExists checks if one or more commands are registered.
func (r *Redis) CommandExists(cmds ...string) bool {
	for _, cmd := range cmds {
		_, ex := r.GetCommands()[cmd]
		if !ex {
			return false
		}
	}
	return true
}

// GetCommandHandler returns the CommandHandler of cmd.
func (r *Redis) GetCommandHandler(cmd string) CommandHandler {
	cmds := r.GetCommands()
	return cmds[cmd]
}

// FlushCommands removes all commands.
func (r *Redis) FlushCommands() {
	r.Commands = make(Commands)
}
