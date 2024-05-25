package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

type Client struct {
	redisClient redis.UniversalClient
	conf        *Config
	ctx         context.Context
	sm          sync.RWMutex
}

func NewRedisClient(conf *Config) (*Client, error) {
	ctx := context.Background()
	redisClient, err := connect(ctx, conf)
	if err != nil {
		return nil, err
	}

	c := &Client{
		conf:        conf,
		ctx:         ctx,
		redisClient: redisClient, // 首次初始化不用加锁
	}
	go c.monitoring() // 监控

	return c, nil
}

func (c *Client) GetRedisClient() redis.UniversalClient {
	c.sm.RLock()
	defer c.sm.RUnlock()

	return c.redisClient
}

// monitoring 重连监控,无限循环，检查redis客服端是否断开连接，如果断开重新连接
func (c *Client) monitoring() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			return
		}
	}()

	for {
		// 先休眠30秒
		time.Sleep(30 * time.Second)
		if c.ping() {
			continue
		}

		fmt.Println("redis 异常断开，正在尝试重连~~~~~")
		c.reConnect()
	}
}

func (c *Client) ping() bool {
	_, err := c.redisClient.Ping(c.ctx).Result()
	return err == nil
}

// reConnect 重连
func (c *Client) reConnect() {
	redisClient, err := connect(c.ctx, c.conf)
	if err != nil {
		fmt.Println("redis 重连失败...")
	}

	// 尝试关闭历史连接
	c.redisClient.Close()

	c.sm.Lock()
	defer c.sm.Unlock()
	c.redisClient = redisClient
}

// connect 建立redis连接
func connect(ctx context.Context, conf *Config) (*redis.Client, error) {
	opt, err := conf.build()
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(opt)
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	// 设置 redisClient name(app name)，方便定位问题
	if err = redisClient.Process(ctx, redis.NewStringCmd(ctx, "client", "setname", fmt.Sprintf("%s", conf.name))); err != nil {
		return nil, err
	}

	return redisClient, nil
}
