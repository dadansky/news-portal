package main

import (
	"log"
	"time"

	redigo "github.com/garyburd/redigo/redis"
)

type Pool interface {
	Get() redigo.Conn
}

type Cache struct {
	Enable          bool
	MaxIdle         int
	MaxActive       int
	IdleTimeoutSecs int
	Address         string
	Auth            string
	DB              string
	Pool            *redigo.Pool
}

// return a new instance of redis pool
func (cache *Cache) NewCachePool() *redigo.Pool {
	if cache.Enable {
		pool := &redigo.Pool{
			MaxIdle:     cache.MaxIdle,
			MaxActive:   cache.MaxActive,
			IdleTimeout: time.Second * time.Duration(cache.IdleTimeoutSecs),
			Dial: func() (redigo.Conn, error) {
				c, err := redigo.Dial("tcp", cache.Address)
				if err != nil {
					return nil, err
				}
				// if _, err := c.Do("AUTH", cache.Auth); err != nil {
				// 	c.Close()
				// 	return nil, err
				// }
				if _, err = c.Do("SELECT", cache.DB); err != nil {
					c.Close()
					return nil, err
				}
				return c, err
			},
			TestOnBorrow: func(c redigo.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		}

		// test connection during init
		c := pool.Get()
		if _, err := c.Do("PING"); err != nil {
			log.Fatal("Cannot connect to Redis: ", err)
		}
		return pool
	}

	return nil
}

func (cache *Cache) getValue(key interface{}) (string, error) {
	if cache.Enable {
		conn := cache.Pool.Get()
		defer conn.Close()
		value, err := redigo.String(conn.Do("GET", key))
		return value, err
	}
	return "", nil
}

func (cache *Cache) setValue(key interface{}, value interface{}) error {
	if cache.Enable {
		conn := cache.Pool.Get()
		defer conn.Close()
		_, err := redigo.String(conn.Do("SET", key, value))
		return err
	}
	return nil
}

func (cache *Cache) enqueueValue(queue string, uuid int) error {
	if cache.Enable {
		conn := cache.Pool.Get()
		defer conn.Close()
		_, err := conn.Do("RPUSH", queue, uuid)
		return err
	}
	return nil
}
