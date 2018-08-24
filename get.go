package redis

import (
	"fmt"
	"github.com/redis-go/redcon"
)

// TODO
func Get(c *Client, cmd redcon.Command) {
	key := string(cmd.Args[1])

	db := c.Redis().RedisDb(c.Db())
	// TODO expired
	//if store.GetDevStore().CheckExpire(&key) != 0 {
	//	c.Conn().WriteNull()
	//	return
	//}

	i := db.Get(&key)
	if i == nil {
		c.Conn().WriteNull()
		return
	}
	if i.ValueType() != "string" {
		c.Conn().WriteError(fmt.Sprintf("%s: key is a %s not a %s", WrongTypeErr, i.ValueType(), "string"))
		return
	}

	v := *i.Value().(*string)

	c.Conn().WriteBulkString(v)
}
