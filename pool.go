package goredis

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"sync/atomic"
	"time"
)

var (
	emptyRedisConn = &RedisConn{
		err: fmt.Errorf("[redis]Error 0003 : RedisPool Get timeout"),
	}
	closedRedisConn = &RedisConn{
		err: fmt.Errorf("[redis]Error 0002 : this pool has been closed"),
	}
)

type Pool struct {
	callback    func() (redis.Conn, error)
	elems       chan *RedisConn
	maxIdle     int32
	maxActive   int32
	curActive   int32
	elemsSize   int32
	status      int32 //1-closed
	timerStatus int32
	waitTime    int   //time for wait
	lifeTime    int32 //time for life,if timeout,will be close the redundant conn
	pingTime    int32 //time for ping
}

func NewPool(callback func() (redis.Conn, error), maxIdle, maxActive int32) *Pool {
	pool := &Pool{
		callback:  callback,
		elems:     make(chan *RedisConn, maxActive),
		maxIdle:   maxIdle,
		maxActive: maxActive,
		waitTime:  0,
		lifeTime:  10,
		pingTime:  20,
	}
	go pool.timerEvent()

	return pool
}

func (this *Pool) SetWaitTime(d int) {
	this.waitTime = d
}

func (this *Pool) SetLifeTime(d int) {
	atomic.StoreInt32(&this.lifeTime, int32(d))
}

func (this *Pool) SetPingTime(d int) {
	this.pingTime = int32(d)
}

func (this *Pool) Update(maxIdle, maxActive int32) {

	if maxIdle == this.maxIdle && maxActive == this.maxActive {
		return
	}
	this.maxIdle = maxIdle
	elems := this.elems
	this.elems = make(chan *RedisConn, maxActive)
	atomic.StoreInt32(&this.elemsSize, 0)
	atomic.StoreInt32(&this.curActive, 0)
	flag := true
	for flag {
		select {
		case e := <-elems:
			select {
			case this.elems <- e:
				atomic.AddInt32(&this.elemsSize, 1)
				atomic.StoreInt32(&this.curActive, 1)
			default:
				flag = false
			}
		default:
			flag = false
		}
	}
	atomic.StoreInt32(&this.maxActive, maxActive)
}

func (this *Pool) TestConn() error {
	conn := this.Get()
	defer conn.Close()
	return conn.Err()
}

func (this *Pool) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	c := this.Get()
	defer c.Close()
	return c.Do(commandName, args...)
}

func (this *Pool) Put(elem *RedisConn) {
	if atomic.LoadInt32(&this.status) != 0 {
		elem.pool = nil
		elem.Close()
		return
	}

	if elem.Err() != nil {
		atomic.AddInt32(&this.curActive, -1)
		elem.pool = nil
		elem.Close()
		return
	}

	select {
	case this.elems <- elem:
		atomic.AddInt32(&this.elemsSize, 1)
		break
	default:
		elem.pool = nil
		elem.Close()
		atomic.AddInt32(&this.curActive, -1)
	}
}

func (this *Pool) Get() *RedisConn {
	var (
		elem *RedisConn
	)
	for {
		elem = this.get()
		if elem.conn != nil && elem.conn.Err() != nil {
			atomic.AddInt32(&this.curActive, -1)
			elem.conn.Close()
			continue
		}
		break
	}

	return elem
}

func (this *Pool) get() *RedisConn {
	if atomic.LoadInt32(&this.status) != 0 {
		return closedRedisConn
	}
	var (
		conn *RedisConn
	)
	select {
	case e := <-this.elems:
		conn = e
		atomic.AddInt32(&this.elemsSize, -1)
	default:
		ca := atomic.LoadInt32(&this.curActive)
		if ca < this.maxActive {
			c, err := this.callback()
			if err != nil {
				conn = &RedisConn{err: err}
				break
			}

			conn = &RedisConn{
				conn: c,
				pool: this,
			}
			atomic.AddInt32(&this.curActive, 1)
		} else {
			fmt.Println("[Warn] 0001 : too many active conn, maxActive=", this.maxActive)
			if this.waitTime != 0 {
				select {
				case conn = <-this.elems:
					atomic.AddInt32(&this.elemsSize, -1)
				case <-time.After(time.Second * time.Duration(this.waitTime)):
					conn = emptyRedisConn
				}
			} else {
				conn = <-this.elems
			}
		}

	}

	return conn
}

func (this *Pool) Close() {
	atomic.StoreInt32(&this.status, 1)
	for {
		select {
		case e := <-this.elems:
			e.pool = nil
			e.Close()
		default:
			return
		}
	}
}

func (this *Pool) timerEvent() {
	timer := time.NewTicker(time.Second * 1)
	defer timer.Stop()
	for atomic.LoadInt32(&this.status) == 0 {

		select {
		case <-timer.C:
			if atomic.LoadInt32(&this.elemsSize) > this.maxIdle {
				this.timerStatus++
				if this.timerStatus > atomic.LoadInt32(&this.lifeTime) {
					select {
					case e := <-this.elems:
						atomic.AddInt32(&this.curActive, -1)
						atomic.AddInt32(&this.elemsSize, -1)
						e.pool = nil
						e.Close()
					default:
						this.timerStatus = 0
					}
				} else {
					this.timerStatus = 0
				}
			}
			flag := true
			n := int(this.elemsSize/this.pingTime + 1)
			for i := 0; (i < n) && flag; i++ {
				select {
				case e := <-this.elems:
					atomic.AddInt32(&this.elemsSize, -1)
					e.Do("PING")
					e.Close()
				default:
					flag = false
					break
				}
			}
		}
	}
}
