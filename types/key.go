package types

import (
	"github.com/redis-go/redis/store"
	"time"
)

type Key struct {
	K *string
	V *string

	Exs bool // can exp
	Ex  time.Time
}

func NewKey(k *string, v *string, exp bool, ex time.Time) *Key {
	return &Key{K: k, V: v, Exs: exp, Ex: ex}
}

func (k *Key) Key() *string {
	return k.K
}

func (k *Key) Value() interface{} {
	return k.V
}

func (k *Key) ValueType() string {
	return "string"
}

func (k *Key) Expire() time.Time {
	return k.Ex
}
func (k *Key) Expires() bool {
	return k.Exs
}

func (k *Key) Expired() bool {
	return k.Expires() && k.Expire().Sub(time.Now()) <= 0*time.Nanosecond
}

func (k *Key) OnDelete(is store.ItemStore) bool {
	return true
}
