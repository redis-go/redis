package redis

import (
	"fmt"
	"github.com/redis-go/redcon"
)

func GetCommand(c *Client, cmd redcon.Command) {
	key := string(cmd.Args[1])

	i := c.Db().GetOrExpire(&key, true)
	if i == nil {
		c.Conn().WriteNull()
		return
	}

	if i.ValueType() != StringType {
		c.Conn().WriteError(fmt.Sprintf("%s: key is a %s not a %s", WrongTypeErr, i.ValueTypeFancy(), StringTypeFancy))
		return
	}

	v := *i.Value().(*string)
	c.Conn().WriteBulkString(v)
}
