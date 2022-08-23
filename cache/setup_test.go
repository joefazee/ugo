package cache

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/dgraph-io/badger/v3"
	"github.com/gomodule/redigo/redis"
	"log"
	"os"
	"testing"
	"time"
)

var testRedisCache RedisCache
var testBadgerCache BadgerCache

func TestMain(m *testing.M) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	pool := redis.Pool{
		MaxActive:   1000,
		MaxIdle:     50,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", s.Addr())
		},
	}

	testRedisCache.Conn = &pool
	testRedisCache.Prefix = "test-ug"

	defer testRedisCache.Conn.Close()

	_ = os.RemoveAll("./testdata/tmp/badger")

	// create a badger database
	if _, err := os.Stat("./testdata/tmp"); os.IsNotExist(err) {
		err := os.Mkdir("./testdata/tmp", 0755)
		if err != nil {
			log.Fatalln(err)
		}
	}

	err = os.Mkdir("./testdata/tmp/badger", 0755)
	if err != nil {
		log.Fatalln(err)
	}

	db, _ := badger.Open(badger.DefaultOptions("./testdata/tmp/badger"))
	testBadgerCache.Conn = db

	code := m.Run()
	_ = os.RemoveAll("./testdata/tmp/badger")

	os.Exit(code)
}
