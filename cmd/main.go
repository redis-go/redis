package main

import (
	"github.com/redis-go/redis/pkg"
	"log"
)

func main() {
	log.Println("Redis started...")
	err := redis.Run()
	if err != nil {
		log.Fatal(err)
	}
}
