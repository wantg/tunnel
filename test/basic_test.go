package test

import (
	"log"
	"testing"

	"github.com/gomodule/redigo/redis"
)

func loadData() {
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Println(err)
	}
	defer c.Close()
	c.Do("SET", "tunnel-test-key", 1)
	origin, _ := redis.Int(c.Do("GET", "tunnel-test-key"))
	new, _ := redis.Int(c.Do("INCR", "tunnel-test-key"))
	log.Println(origin, new)
}

func TestTunnel(t *testing.T) {
	count := 500
	for idx := 0; idx < count; idx++ {
		log.Println()
		log.Println(idx)
		loadData()
	}
}
