package cmds

import (
	"fmt"
	"github.com/redis-go/redcon"
	"github.com/redis-go/redis"
	"github.com/redis-go/redis/store"
	"github.com/redis-go/redis/types"
	"strconv"
	"strings"
	"time"
)

func Set(c redcon.Conn, cmd redcon.Command, _ *redis.Redis) {
	if len(cmd.Args) == 1 { // nothing done
		c.WriteString("OK")
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
				c.WriteError(redis.SyntaxERR)
				return
			case "ex":
				if isPX { // is already px
					c.WriteError(redis.SyntaxERR)
					return
				}

				// was last arg?
				if len(cmd.Args) == i {
					c.WriteError(redis.SyntaxERR)
					return
				}

				// read next arg
				i++
				i, err := strconv.ParseUint(string(cmd.Args[i]), 10, 64)
				if err != nil {
					c.WriteError(fmt.Sprintf("%s: %s", redis.InvalidIntErr, err.Error()))
					return
				}
				if i == 0 {
					c.WriteError("ERR invalid expire time in set: cannot be 0")
					return
				}
				expire = time.Now().Add(time.Duration(i * uint64(time.Second)))
				yesExpire, isEX = true, true
				i++
				continue
			case "px":
				if isEX { // is already ex
					c.WriteError(redis.SyntaxERR)
					return
				}

				// was last arg?
				if len(cmd.Args) == i {
					c.WriteError(redis.SyntaxERR)
					return
				}

				// read next arg
				i++
				i, err := strconv.ParseUint(string(cmd.Args[i]), 10, 64)
				if err != nil {
					c.WriteError(fmt.Sprintf("%s: %s", redis.InvalidIntErr, err.Error()))
					return
				}
				if i == 0 {
					c.WriteError("ERR invalid expire time in set: cannot be 0")
					return
				}
				expire = time.Now().Add(time.Duration(i * uint64(time.Millisecond)))
				yesExpire, isPX = true, true
				i++
				continue
			case "nx":
				if XX { // is already xx
					c.WriteError(redis.SyntaxERR)
					return
				}
				NX = true
				i++
				continue
			case "xx":
				if NX { // is already nx
					c.WriteError(redis.SyntaxERR)
					return
				}
				XX = true
				i++
				continue
			}
		}
	}

	// Only set the key if it does not already exist.
	if NX && store.GetDevStore().ItemExists(key) {
		c.WriteNull()
		return
	} else if XX && !store.GetDevStore().ItemExists(key) { // Only set the key if it already exist.
		c.WriteNull()
		return
	}

	store.GetDevStore().AddItem(key, types.NewKey(key, &value, yesExpire, expire))
	c.WriteString("OK")
}
