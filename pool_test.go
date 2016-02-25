package goredis

import (
	"reflect"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	pool      *Pool
	maxIdle   = int32(16)
	maxActive = int32(1024)
)

func TestNewPool(t *testing.T) {
	pool = NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", "127.0.0.1:6379")
		if err != nil {
			t.Fatal(err)
		}
		return c, err
	},
		maxIdle,
		maxActive)
	if err := pool.TestConn(); err != nil {
		t.Fatal(err)
	}
}

func TestPoolGet(t *testing.T) {
	ch := make([]*RedisConn, 1024)
	var err error
	for i := 0; i < 1024; i++ {
		ch[i] = pool.Get()
		if ch[i].Err() != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 1024; i++ {
		pool.Put(ch[i])
	}
}

func TestPoolDo(t *testing.T) {
	reply, err := pool.Do("SET", "test", "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reply.(string); !ok {
		t.Fatal("wrong reply type|type=", reflect.TypeOf(reply))
	} else if reply.(string) != "OK" {
		t.Fatal("reply wrong|reply=", reply)
	}
	reply, err = pool.Do("GET", "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reply.([]uint8); !ok {
		t.Fatal("wrong reply type|type=", reflect.TypeOf(reply))
	} else if string(reply.([]uint8)) != "test" {
		t.Fatal("reply wrong|reply=", reply)
	}
	pool.Do("DEL", "test")
}

func TestPoolTimerEvent(t *testing.T) {
	pool.Close()
	maxIdle = int32(3)
	maxActive = int32(8)
	pool = NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", "192.168.1.202:6379")
		if err != nil {
			t.Fatal(err)
		}
		return c, err
	},
		maxIdle,
		maxActive)
	pool.SetLifeTime(0)
	elems := make([]*RedisConn, maxActive)
	for i := 0; i < int(maxActive); i++ {
		elems[i] = pool.Get()
		if elems[i].Err() != nil {
			t.Fatal(elems[i].Err())
		}
	}
	if pool.curActive != maxActive {
		t.Fatal("size wrong|curActive=", pool.curActive, "|maxActive=", maxActive)
	}
	for i := 0; i < int(maxActive); i++ {
		elems[i].Close()
	}
	if pool.elemsSize != maxActive {
		t.Fatal("size wrong|elemsSize=", pool.elemsSize, "|maxActive=", maxActive)
	}
	time.AfterFunc(time.Second*3, func() {
		if pool.elemsSize != maxActive-3 || pool.curActive != maxActive-3 {
			t.Fatal("elemsSize=", pool.elemsSize, "|curActive=", pool.curActive)
		}
	})
	time.Sleep(time.Second * 4)
}

func TestPoolWait(t *testing.T) {
	pool.Update(1, 1)
	pool.SetWaitTime(1)
	{
		conn := pool.Get()
		if conn.Err() != nil {
			t.Error(conn.Err())
		}
		time.AfterFunc(time.Millisecond*900, func() { conn.Close() })
	}
	{
		conn := pool.Get()
		if conn.Err() != nil {
			t.Error(conn.Err())
		}
		conn.Close()
	}
	{
		conn := pool.Get()
		if conn.Err() != nil {
			t.Error(conn.Err())
		}
		time.AfterFunc(time.Millisecond*1100, func() { conn.Close() })
	}
	{
		conn := pool.Get()
		defer conn.Close()
		if conn.Err() == nil {
			t.Error("failed")
		}
	}
}

func TestPoolSend(t *testing.T) {
	conn := pool.Get()
	defer conn.Close()
	{
		if err := conn.Send("SET", "SEND", "test"); err != nil {
			t.Error(err)
		}
		if err := conn.Send("GET", "SEND"); err != nil {
			t.Error(err)
		}
		if err := conn.Send("DEL", "SEND"); err != nil {
			t.Error(err)
		}
	}
	{
		if err := conn.Flush(); err != nil {
			t.Error(err)
		}
	}
	{
		rp, err := conn.Receive()
		if err != nil {
			t.Error(err)
		}
		if rp.(string) != "OK" {
			t.Error(rp)
		}
	}
	{
		rp, err := conn.Receive()
		if err != nil {
			t.Error(err)
		}
		if string(rp.([]byte)) != "test" {
			t.Error(rp)
		}
	}
	{
		rp, err := conn.Receive()
		if err != nil {
			t.Error(err)
		}
		if rp.(int64) != 1 {
			t.Error(rp)
		}
	}
}
