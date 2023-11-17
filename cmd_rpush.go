package redis

import (
	"fmt"
	"github.com/redis-go/redcon"
	"time"
)

func RPushCommand(c *Client, cmd redcon.Command) {
	if len(cmd.Args) < 3 {
		c.Conn().WriteError(fmt.Sprintf(WrongNumOfArgsErr, "rpush"))
		return
	}
	key := string(cmd.Args[1])

	db := c.Db()
	i := db.GetOrExpire(&key, true)
	if i == nil {
		i = NewList()
		db.Set(&key, i, false, time.Time{})
	} else if i.Type() != ListType {
		c.Conn().WriteError(fmt.Sprintf("%s: key is a %s not a %s", WrongTypeErr, i.TypeFancy(), ListTypeFancy))
		return
	}

	l := i.(*List)
	var length int
	db.Mu().Lock()
	for j := 2; j < len(cmd.Args); j++ {
		v := string(cmd.Args[j])
		length = l.RPush(&v)
	}
	db.Mu().Unlock()

	c.Conn().WriteInt(length)
}
