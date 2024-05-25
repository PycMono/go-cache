package redis

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// Config 配置文件
type Config struct {
	name         string        // app name
	addr         string        // redis addr，例如 127.0.0.1:6379
	password     string        // redis password
	db           int           // redis db
	poolSize     int           // redis pool size
	poolTimeout  time.Duration // redis 超时（单位秒,默认为0）
	readTimeout  time.Duration // redis 超时（单位秒,默认为0）
	writeTimeout time.Duration // redis 超时（单位秒,默认为0）
}

func (c Config) WithName(name string) Config {
	c.name = name
	return c
}

func (c Config) WithAddr(addr string) Config {
	c.addr = addr
	return c
}

func (c Config) WithPassword(password string) Config {
	c.password = password
	return c
}

func (c Config) WithDB(db int) Config {
	c.db = db
	return c
}

func (c Config) WithPoolSize(poolSize int) Config {
	c.poolSize = poolSize
	return c
}

func (c Config) WithPoolTimeout(poolTimeout time.Duration) Config {
	c.poolTimeout = poolTimeout
	return c
}

func (c Config) WithReadTimeout(readTimeout time.Duration) Config {
	c.readTimeout = readTimeout
	return c
}

func (c Config) WithWriteTimeout(writeTimeout time.Duration) Config {
	c.writeTimeout = writeTimeout
	return c
}

func (c Config) build() (*redis.Options, error) {
	if len(c.name) == 0 {
		return nil, fmt.Errorf("name为空")
	}
	if len(c.addr) == 0 {
		return nil, fmt.Errorf("addr为空")
	}
	if len(c.password) == 0 {
		return nil, fmt.Errorf("password为空")
	}
	if c.db == 0 {
		return nil, fmt.Errorf("db为空")
	}
	if c.poolSize == 0 {
		c.poolSize = 30
	}

	return &redis.Options{
		Addr:         c.addr,
		ClientName:   c.name,
		Password:     c.password,
		DB:           c.db,
		WriteTimeout: c.writeTimeout,
		PoolSize:     c.poolSize,
		PoolTimeout:  c.poolTimeout,
		ReadTimeout:  c.readTimeout,
	}, nil
}
