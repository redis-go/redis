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

	p := client.Pipeline()
	r1 := p.Ping()
	r2 := p.Ping()
	fmt.Println(p.Exec())

	fmt.Println(r1.Val())
	fmt.Println(r2.Val())

	v, err := client.Set("abc", "test", 99*time.Second).Result()
	if err != nil {
		fmt.Println("err:", err.Error())
	} else {
		fmt.Println(v)
	}
	//fmt.Println(client.Get("abc"))
	//fmt.Println(client.Del("abc"))
	//fmt.Println(client.Do("detach"))
}
