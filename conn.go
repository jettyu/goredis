package goredis

import (
	"github.com/garyburd/redigo/redis"
)

type RedisConn struct {
	conn redis.Conn
	pool *Pool
	err  error
	//	status int32
}

func (this *RedisConn) Close() error {
	if this.err != nil || this.pool == nil {
		return nil
	} else if this.conn != nil {
		if this.conn.Err() != nil {
			return this.conn.Close()
		} else {
			this.pool.Put(this)
		}
	}

	return nil
}

func (this *RedisConn) Command(commandName string, args ...interface{}) *RedisReply {
	reply, err := this.Do(commandName, args...)
	return NewRedisReply(reply, err)
}

func (this *RedisConn) Conn() redis.Conn {
	return this.conn
}

func (this *RedisConn) Err() error {
	if this.err != nil {
		return this.err
	}
	return this.conn.Err()
}

func (this *RedisConn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if this.err != nil {
		return nil, this.err
	}
	return this.conn.Do(commandName, args...)
}

func (this *RedisConn) Send(commandName string, args ...interface{}) error {
	if this.err != nil {
		return this.err
	}
	return this.conn.Send(commandName, args...)
}

func (this *RedisConn) Flush() error {
	if this.err != nil {
		return this.err
	}
	return this.conn.Flush()
}

func (this *RedisConn) Receive() (reply interface{}, err error) {
	if this.err != nil {
		return nil, this.err
	}
	return this.conn.Receive()
}
