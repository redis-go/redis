package redis

import (
	"fmt"
	"github.com/redis-go/redcon"
	"strconv"
)

/*
Removes the first count occurrences of elements equal to element from the list stored at key. The count argument influences the operation in the following ways:

count > 0: Remove elements equal to element moving from head to tail.
count < 0: Remove elements equal to element moving from tail to head.
count = 0: Remove all elements equal to element.
For example, LREM list -2 "hello" will remove the last two occurrences of "hello" in the list stored at list.

Note that non-existing keys are treated like empty lists, so when key does not exist, the command will always return 0.
*/
func LRemCommand(c *Client, cmd redcon.Command) {
	if len(cmd.Args) < 3 {
		c.Conn().WriteError(fmt.Sprintf(WrongNumOfArgsErr, "lrem"))
		return
	}

	key := string(cmd.Args[1])
	count, err := strconv.Atoi(string(cmd.Args[2]))
	if err != nil {
		c.Conn().WriteError(fmt.Sprintf(InvalidIntErr))
		return
	}

	db := c.Db()
	var removed int

	i := db.GetOrExpire(&key, false)
	if i == nil {
		c.Conn().WriteInt(0)
		return
	}

	l, ok := i.(*List)
	if !ok {
		c.Conn().WriteError(fmt.Sprintf(WrongTypeErr))
		return
	}

	var toDelete bool
	func() {
		db.Mu().Lock()
		defer db.Mu().Unlock()

		if count > 0 {
			for j := 0; j < count; j++ {
				v, b := l.LPop()
				if b {
					toDelete = true
					break
				}
				if v == nil {
					break
				}
				removed++
			}
		} else if count < 0 {
			for j := 0; j < -count; j++ {
				v, b := l.RPop()
				if b {
					toDelete = true
					break
				}
				if v == nil {
					break
				}
				removed++
			}
		} else {
			for {
				v, b := l.LPop()
				if b {
					toDelete = true
					break
				}
				if v == nil {
					break
				}
				removed++
			}
		}
	}()
	if toDelete {
		db.Delete(&key)
	}

	fmt.Println("LREM", key, count, removed)
}
