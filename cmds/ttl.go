package cmds

import (
	"fmt"
	"github.com/redis-go/redcon"
	"github.com/redis-go/redis"
	"github.com/redis-go/redis/store"
	"time"
)

func Ttl(c redcon.Conn, cmd redcon.Command, _ *redis.Redis) {
	if len(cmd.Args) != 2 {
		c.WriteError(fmt.Sprintf("wrong number of arguments (given %d, expected 1)", len(cmd.Args)-1))
		return
	}

	key := string(cmd.Args[1])
	i, ex := store.GetDevStore().GetItem(&key)
	if !ex {
		c.WriteInt(-2)
		return
	} else if !i.Expires() {
		c.WriteInt(-1)
		return
	}

	// TODO delete before, if expired
	c.WriteInt64(int64(time.Now().Sub(i.Expire()).Seconds()))
}
