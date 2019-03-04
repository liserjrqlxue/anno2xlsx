package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
)

func main() {
	redisdb := redis.NewClient(&redis.Options{
		Addr:     "192.168.136.114:6380", // use default Addr
		Password: "",                     // no password set
		DB:       0,                      // use default DB
	})

	pong, err := redisdb.Ping().Result()
	fmt.Println(pong, err)

	//var v=redisdb.HGet("nm2ensp","XM_003906460")
	var v = redisdb.HGet("nm2ensp", "XM_003906460xx")
	fmt.Println(v)
	fmt.Println(v.Args())
	fmt.Println(v.Result())
	r, e := v.Result()
	fmt.Println("result:", r)
	fmt.Println("error:", e)

	v = redisdb.HGet("SEQ500_all_native_snp", "chr1_200567334_C_A")
	fmt.Println(v)
	fmt.Println(v.Args())
	fmt.Println(v.Result())
	r, e = v.Result()
	fmt.Println("result:", r)
	fmt.Println("error:", e)
	var rs []string
	err = json.Unmarshal([]byte(r), &rs)
	fmt.Println(err)
	fmt.Println("result:", rs)
	fmt.Println("result:", rs[1])

}
