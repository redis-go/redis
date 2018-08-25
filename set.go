package redis

import (
	"fmt"
	"github.com/redis-go/redcon"
	"github.com/redis-go/redis/types"
	"strconv"
	"strings"
	"time"
)

func SetCommand(c *Client, cmd redcon.Command) {
	if len(cmd.Args) == 1 { // nothing done
		c.Conn().WriteString("OK")
		return
	}

	k := string(cmd.Args[1])
	key := &k
	var value string
	if len(cmd.Args) > 1 {
		value = string(cmd.Args[2])
	}

	var yesExpire bool
	var expire time.Time

	var isEX bool
	var isPX bool

	var NX bool
	var XX bool

	if len(cmd.Args) > 2 {
		for i := 3; i+1 < len(cmd.Args); {
			arg := strings.ToLower(string(cmd.Args[i]))
			switch arg {
			default:
				c.Conn().WriteError(SyntaxERR)
				return
			case "ex":
				if isPX { // is already px
					c.Conn().WriteError(SyntaxERR)
					return
				}

				// was last arg?
				if len(cmd.Args) == i {
					c.Conn().WriteError(SyntaxERR)
					return
				}

				// read next arg
				i++
				i, err := strconv.ParseUint(string(cmd.Args[i]), 10, 64)
				if err != nil {
					c.Conn().WriteError(fmt.Sprintf("%s: %s", InvalidIntErr, err.Error()))
					return
				}
				if i == 0 {
					c.Conn().WriteError("ERR invalid expire time in set: cannot be 0")
					return
				}
				expire = time.Now().Add(time.Duration(i * uint64(time.Second)))
				yesExpire, isEX = true, true
				i++
				continue
			case "px":
				if isEX { // is already ex
					c.Conn().WriteError(SyntaxERR)
					return
				}

				// was last arg?
				if len(cmd.Args) == i {
					c.Conn().WriteError(SyntaxERR)
					return
				}

				// read next arg
				i++
				i, err := strconv.ParseUint(string(cmd.Args[i]), 10, 64)
				if err != nil {
					c.Conn().WriteError(fmt.Sprintf("%s: %s", InvalidIntErr, err.Error()))
					return
				}
				if i == 0 {
					c.Conn().WriteError("ERR invalid expire time in set: cannot be 0")
					return
				}
				expire = time.Now().Add(time.Duration(i * uint64(time.Millisecond)))
				yesExpire, isPX = true, true
				i++
				continue
			case "nx":
				if XX { // is already xx
					c.Conn().WriteError(SyntaxERR)
					return
				}
				NX = true
				i++
				continue
			case "xx":
				if NX { // is already nx
					c.Conn().WriteError(SyntaxERR)
					return
				}
				XX = true
				i++
				continue
			}
		}
	}

	// clients selected db
	db := c.Redis().RedisDb(c.Db())

	exists := db.Exists(key)
	if NX && exists || XX && !exists {
		c.Conn().WriteNull()
		return
	}

	db.Set(key, types.NewString(&value, yesExpire, expire))
	c.Conn().WriteString("OK")
}
