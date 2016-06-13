package main

import (
	"fmt"
	redis "github.com/garyburd/redigo/redis"
	"github.com/jettyu/goredis"
)

func TestPool() {
	ri := goredis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", "127.0.0.1:6379")
		fmt.Println("Dial ...")
		if err != nil {
			fmt.Println(err)
		}
		return c, err
	},
		8,
		8)
	conn := ri.Get()
	defer conn.Close()
	reply, err := conn.Do("SET", "test", "hello")
	fmt.Println(reply, err)
	reply, err = ri.Do("GET", "test")
	fmt.Println(reply, err)
	rp := conn.Command("GET", "test")
	fmt.Println(rp)
}

func TestCluster() {
	redishosts := []string{
		"127.0.0.1:6380",
		"127.0.0.1:6381",
		"127.0.0.1:6382",
		"127.0.0.1:6383",
		"127.0.0.1:6384",
		"127.0.0.1:6385",
	}
	ri := goredis.NewRedisCluster(redishosts, 8, 64, true)
	reply, err := ri.Do("SET", "test", "hello")
	fmt.Println(reply, err)
	reply, err = ri.Do("GET", "test")
	fmt.Println(reply, err)
}

func main() {
	TestPool()
	//TestCluster()
}
