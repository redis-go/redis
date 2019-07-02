package main

import (
	"github.com/redis-go/redis/pkg"
)

func main() {
	r := redis.NewRedis()
	if err := r.Run(); err != nil {
		r.Logger().Sugar().Error(err)
	}
}
