package goredis

import (
	"reflect"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	_testPool      *Pool
	maxIdle   = int32(16)
	maxActive = int32(1024)
)

func TestNewPool(t *testing.T) {
	_testPool = NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", "127.0.0.1:6379")
		if err != nil {
			t.Fatal(err)
		}
		return c, err
	},
		maxIdle,
		maxActive)
	if err := _testPool.TestConn(); err != nil {
		t.Fatal(err)
	}
}

func TestPoolGet(t *testing.T) {
	ch := make([]*RedisConn, 1024)
	var err error
	for i := 0; i < 1024; i++ {
		ch[i] = _testPool.Get()
		if ch[i].Err() != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 1024; i++ {
		_testPool.Put(ch[i])
	}
}

func TestPoolDo(t *testing.T) {
	reply, err := _testPool.Do("SET", "test", "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reply.(string); !ok {
		t.Fatal("wrong reply type|type=", reflect.TypeOf(reply))
	} else if reply.(string) != "OK" {
		t.Fatal("reply wrong|reply=", reply)
	}
	reply, err = _testPool.Do("GET", "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := reply.([]uint8); !ok {
		t.Fatal("wrong reply type|type=", reflect.TypeOf(reply))
	} else if string(reply.([]uint8)) != "test" {
		t.Fatal("reply wrong|reply=", reply)
	}
	_testPool.Do("DEL", "test")
}

func TestPoolTimerEvent(t *testing.T) {
	_testPool.Close()
	maxIdle = int32(3)
	maxActive = int32(8)
	_testPool = NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", "192.168.1.202:6379")
		if err != nil {
			t.Fatal(err)
		}
		return c, err
	},
		maxIdle,
		maxActive)
	_testPool.SetLifeTime(0)
	elems := make([]*RedisConn, maxActive)
	for i := 0; i < int(maxActive); i++ {
		elems[i] = _testPool.Get()
		if elems[i].Err() != nil {
			t.Fatal(elems[i].Err())
		}
	}
	if _testPool.curActive != maxActive {
		t.Fatal("size wrong|curActive=", _testPool.curActive, "|maxActive=", maxActive)
	}
	for i := 0; i < int(maxActive); i++ {
		elems[i].Close()
	}
	if _testPool.elemsSize != maxActive {
		t.Fatal("size wrong|elemsSize=", _testPool.elemsSize, "|maxActive=", maxActive)
	}
	time.AfterFunc(time.Second*3, func() {
		if _testPool.elemsSize != maxActive-3 || _testPool.curActive != maxActive-3 {
			t.Fatal("elemsSize=", _testPool.elemsSize, "|curActive=", _testPool.curActive)
		}
	})
	time.Sleep(time.Second * 4)
}

func TestPoolWait(t *testing.T) {
	_testPool.Update(1, 1)
	_testPool.SetWaitTime(1)
	{
		conn := _testPool.Get()
		if conn.Err() != nil {
			t.Error(conn.Err())
		}
		time.AfterFunc(time.Millisecond*900, func() { conn.Close() })
	}
	{
		conn := _testPool.Get()
		if conn.Err() != nil {
			t.Error(conn.Err())
		}
		conn.Close()
	}
	{
		conn := _testPool.Get()
		if conn.Err() != nil {
			t.Error(conn.Err())
		}
		time.AfterFunc(time.Millisecond*1100, func() { conn.Close() })
	}
	{
		conn := _testPool.Get()
		defer conn.Close()
		if conn.Err() == nil {
			t.Error("failed")
		}
	}
}

func TestPoolSend(t *testing.T) {
	conn := _testPool.Get()
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
