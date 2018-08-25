package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/hako/durafmt"
	"math/rand"
	"time"
)

func main() {
	ExampleNewClient()
}

func ExampleNewClient() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	for i := 0; i < 100000; i++ {
		go func() {
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			start := time.Now()
			client.Set("milli", "value", 0).Result()
			if d := durafmt.Parse(time.Now().Sub(start)).String(); d != "" { //&& d != "0 seconds" {
				fmt.Println(d)
			}
		}()
		time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
	}
	fmt.Println("done")

	time.Sleep(5 * time.Hour)
}
