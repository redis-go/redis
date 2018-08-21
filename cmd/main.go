package main

import (
	"github.com/redis-go/redis"
	"log"
)

func main() {
	log.Println("Work in Progress")
	log.Fatal(redis.Run(":6379"))
}
