package cmds

import (
	"bytes"
	"fmt"
	"github.com/redis-go/redcon"
	"github.com/redis-go/redis"
)

func Ping(c redcon.Conn, cmd redcon.Command, _ *redis.Redis) {
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
}
