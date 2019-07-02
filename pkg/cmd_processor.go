package redis

import "go.uber.org/zap"

type cmdProcessor struct {
	cmds   map[string]*Command
	logger *zap.Logger
}

func newCmdProcessor(logger *zap.Logger) *cmdProcessor {
	return &cmdProcessor{
		cmds:   make(map[string]*Command),
		logger: logger.With(zap.String("sector", "cmdProcessor")),
	}
}

type Command struct {
	name string
	run  cmdRun
}

type cmdRun func(c *client, args [][]byte)

// run
func (p *cmdProcessor) run(name string, c *client, args [][]byte) {
	cmd, ok := p.cmds[name]
	if ok {
		cmd.run(c, args)
	}
}

// registers command, returns false if command name already used
func (p *cmdProcessor) addCommand(cmd *Command) bool {
	_, ok := p.cmds[cmd.name]
	if !ok {
		p.cmds[cmd.name] = cmd
	}
	return !ok
}

func (p *cmdProcessor) newCmd(name string, run cmdRun) *Command {
	return &Command{
		name: name,
		run:  run,
	}
}
