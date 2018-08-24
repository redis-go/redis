package redis

import (
	"fmt"
	"github.com/redis-go/redcon"
	"time"
)

// TODO
func Ttl(c *Client, cmd redcon.Command) {
	if len(cmd.Args) != 2 {
		c.Conn().WriteError(fmt.Sprintf("wrong number of arguments (given %d, expected 1)", len(cmd.Args)-1))
		return
	}

	key := string(cmd.Args[1])
	i := c.Redis().RedisDb(c.Db()).Get(&key)
	if i == nil {
		c.Conn().WriteInt(-2)
		return
	} else if !i.Expires() {
		c.Conn().WriteInt(-1)
		return
	}

	// TODO delete before, if expired
	c.Conn().WriteInt64(int64(time.Now().Sub(i.Expiry()).Seconds()))
}
