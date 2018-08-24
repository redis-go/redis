package cmds

import (
	"fmt"
	"github.com/redis-go/redcon"
	"github.com/redis-go/redis"
	"github.com/redis-go/redis/store"
)

func Get(c redcon.Conn, cmd redcon.Command, _ *redis.Redis) {
	key := string(cmd.Args[1])

	// expired
	if store.GetDevStore().CheckExpire(&key) != 0 {
		c.WriteNull()
		return
	}

	i, ok := store.GetDevStore().GetItem(&key)
	if !ok {
		c.WriteNull()
		return
	}
	if i.ValueType() != "string" {
		c.WriteError(fmt.Sprintf("%s: key is a %s not a %s", redis.WrongTypeErr, i.ValueType(), "string"))
		return
	}

	v := *i.Value().(*string)

	c.WriteBulkString(v)
}
