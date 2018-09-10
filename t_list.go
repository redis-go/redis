package redis

import (
	"time"
)

const ListType = uint64(0)
const ListTypeFancy = "list"

var _ Item = (*List)(nil)

type List struct {
	elements map[string]struct{}
}

func (l *List) Value() interface{} {
	return l.elements
}

func (l *List) ValueType() uint64 {
	return ListType
}

func (l *List) ValueTypeFancy() string {
	return ListTypeFancy
}

func (l *List) Expiry() time.Time {
	panic("implement me")
}

func (l *List) Expires() bool {
	panic("implement me")
}

func (l *List) OnDelete(key *string, db *RedisDb) {
	panic("implement me")
}
