package cmds

import (
	"github.com/redis-go/redcon"
	"github.com/redis-go/redis"
	"github.com/redis-go/redis/store"
)

func Del(c redcon.Conn, cmd redcon.Command, _ *redis.Redis) {
	key := string(cmd.Args[1])
	store.Store.DeleteItem(&key)
	c.WriteInt(1)
}
