package redis

import (
	"context"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	conf := Config{}.WithName("test").
		WithDB(10).
		WithAddr(":6379").
		WithPassword("123456").
		WithPoolSize(100)
	redisClient, err := NewRedisClient(&conf)
	if err != nil {
		panic(err)
	}

	var (
		val = "22222222"
		m   = make(map[string][]byte)
	)
	m["12pyc1222"] = []byte(val)

	redisAdaptor := NewRedisAdaptor(redisClient)
	err = redisAdaptor.Set(context.TODO(), m, time.Hour)
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Hour)
}
