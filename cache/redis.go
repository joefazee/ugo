package cache

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
)

type RedisCache struct {
	Conn   *redis.Pool
	Prefix string
}

func (c *RedisCache) key(str string) string {
	return fmt.Sprintf("%s:%s", c.Prefix, str)
}

func (c *RedisCache) Get(key string) (interface{}, error) {
	key = c.key(key)
	conn := c.Conn.Get()
	defer conn.Close()

	cacheEntry, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}

	decoded, err := decode(string(cacheEntry))
	if err != nil {
		return nil, err
	}

	item := decoded[key]

	return item, nil
}

func (c *RedisCache) Set(key string, value interface{}, expires ...int) error {
	key = c.key(key)
	conn := c.Conn.Get()
	defer conn.Close()

	var err error

	entry := Entry{}
	entry[key] = value
	encoded, err := encode(entry)
	if err != nil {
		return nil
	}

	if len(expires) > 0 {
		_, err = conn.Do("SETEX", key, expires[0], string(encoded))
	} else {
		_, err = conn.Do("SET", key, string(encoded))
	}
	return err
}

func (c *RedisCache) Forget(key string) error {
	key = c.key(key)
	conn := c.Conn.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)

	return err

}

func (c *RedisCache) EmptyByMatch(key string) error {
	pattern := fmt.Sprintf("%s*", c.key(key))
	return c.forgetByKeys(pattern)
}

func (c *RedisCache) forgetByKeys(pattern string) error {

	conn := c.Conn.Get()
	defer conn.Close()

	keys, err := c.getKeys(pattern)
	if err != nil {
		return err
	}

	for _, x := range keys {
		_, err := conn.Do("DEL", x)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *RedisCache) Empty() error {
	key := c.key("*")
	return c.forgetByKeys(key)
}

type Entry map[string]interface{}

func (c *RedisCache) Has(key string) (bool, error) {
	key = fmt.Sprintf("%s:%s", c.Prefix, key)
	conn := c.Conn.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (c *RedisCache) getKeys(pattern string) ([]string, error) {

	conn := c.Conn.Get()
	defer conn.Close()

	iter := 0
	var keys []string
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return keys, err
		}

		iter, _ := redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
}
