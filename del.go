package redis

import (
	"github.com/redis-go/redcon"
)

// TODO
func Del(c *Client, cmd redcon.Command) {
	key := string(cmd.Args[1])
	c.Redis().RedisDb(c.Db()).Delete(&key)
	c.Conn().WriteInt(1)
}
