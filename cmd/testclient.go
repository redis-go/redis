package main

import (
	"fmt"
	"github.com/go-redis/redis"
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

	for i := 0; i < 1; i++ {
		go func() {
			vs, err := client.Set("milli", "value", 3*time.Millisecond).Result()
			if err != nil {
				fmt.Println("err:", err.Error())
			} else {
				fmt.Println(vs)
			}
			time.Sleep(100 * time.Millisecond)
			vs, err = client.Get("milli").Result() // passive expire currently not impl so expect value
			if err != nil {
				fmt.Println("err:", err.Error())
			} else {
				fmt.Println(vs)
			}
		}()
	}

	time.Sleep(5 * time.Second)
}
