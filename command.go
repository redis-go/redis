package redis

import "github.com/redis-go/redcon"

// Commands map
type Commands map[string]CommandHandler

// The CommandHandler is triggered when the received
// command equals a registered command.
//
// However the CommandHandler is executed by the Handler,
// so if you implement an own Handler make sure the CommandHandler is called.
type CommandHandler func(c *Client, cmd redcon.Command)

// Is called when a request is received,
// after Accept and if the command is not registered.
//
// However UnknownCommand is executed by the Handler,
// so if you implement an own Handler make sure to include UnknownCommand.
type UnknownCommand func(c *Client, cmd redcon.Command)

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

// FlushCommands removes all commands.
func (r *Redis) FlushCommands() {
	r.Mu().Lock()
	defer r.Mu().Unlock()
	r.commands = make(Commands)
}

// CommandHandlerFn returns the CommandHandler of cmd.
func (r *Redis) CommandHandlerFn(cmd string) CommandHandler {
	r.Mu().RLock()
	defer r.Mu().RUnlock()
	return r.getCommands()[cmd]
}

// UnknownCommandFn returns the UnknownCommand function.
func (r *Redis) UnknownCommandFn() UnknownCommand {
	r.Mu().RLock()
	defer r.Mu().RUnlock()
	return r.unknownCommand
}
