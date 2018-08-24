package redis

import (
	"bytes"
	"fmt"
	"github.com/redis-go/redcon"
)

func Ping(c *Client, cmd redcon.Command) {
	if len(cmd.Args) > 1 {
		var buf bytes.Buffer
		for i := 1; i-1 < len(cmd.Args); i++ {
			buf.Write(cmd.Args[i])
			buf.WriteString(" ")
			fmt.Println(i)
		}
		s := buf.String()
		s = s[:len(s)-1]
		c.Conn().WriteString(s)
		return
	}
	c.Conn().WriteString("PONG")
}
