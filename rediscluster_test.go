package goredis

import (
	"testing"
)

var (
	_testCluster RedisCluster = NewRedisCluster(
		[]string{
			"127.0.0.1:7000",
			"127.0.0.1:7001",
			"127.0.0.1:7002",
			"127.0.0.1:7003",
			"127.0.0.1:7004",
			"127.0.0.1:7005",
		},
		8,
		8,
		false,
	)
)

func TestRedisCluster(t *testing.T) {
	if err := _testCluster.TestCluster(); err != nil {
		t.Fatal(err)
	}
}

func TestClusterDo(t *testing.T) {
	{
		rp, err := _testCluster.Do("SET", "CLUSTER", "test")
		if err != nil {
			t.Fatal(err)
		}
		if rp.(string) != "OK" {
			t.Fatal(rp)
		}
	}
	_testCluster.Do("DEL", "CLUSTER")
}

func TestClusterGetHandle(t *testing.T) {
	rh := _testCluster.GetHandle("CLUSTER")
	if rh == nil {
		t.Fatal("rh is nil")
	}
	conn := rh.Get()
	defer conn.Close()
	if err := conn.Send("SET", "CLUSTER", "test"); err != nil {
		t.Fatal(err)
	}
	if err := conn.Send("GET", "CLUSTER"); err != nil {
		t.Fatal(err)
	}
	if err := conn.Send("DEL", "CLUSTER"); err != nil {
		t.Fatal(err)
	}
	if err := conn.Flush(); err != nil {
		t.Fatal(err)
	}
	{
		replys := make([]interface{}, 3)
		errs := make([]error, 3)
		for i := 0; i < 3; i++ {
			replys[i], errs[i] = conn.Receive()
			if errs[i] != nil {
				t.Fatal(errs[i])
			}
		}
		if replys[0].(string) != "OK" {
			t.Fatal(replys[0])
		}
		if string(replys[1].([]byte)) != "test" {
			t.Fatal(replys[1])
		}
		if replys[2].(int64) != 1 {
			t.Fatal(replys[2])
		}
	}
}
