package redis

import (
	"context"
	"errors"
	"github.com/PycMono/go-cache/client"
	"github.com/avast/retry-go"
	"github.com/redis/go-redis/v9"
	"time"
)

type Cache struct {
	client *Client
}

func NewRedisAdaptor(client *Client) client.IAdaptor {
	return &Cache{client: client}
}

func (r *Cache) Set(ctx context.Context, params map[string][]byte, expire time.Duration) error {
	m := make(map[string]interface{})
	for k, v := range params {
		m[k] = string(v)
	}

	_, err := r.client.GetRedisClient().Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for key, value := range m {
			err := pipe.Set(ctx, key, value, expire).Err()
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (r *Cache) Del(ctx context.Context, k []string) error {
	return retry.Do(
		func() error {
			out := r.client.GetRedisClient().Del(ctx, k...)
			return out.Err()
		},
		retry.RetryIf(func(err error) bool {
			return err != nil
		}),
		retry.Delay(3*time.Second),
		retry.Attempts(3),
	)
}

func (r *Cache) Get(ctx context.Context, k []string) (map[string][]byte, error) {
	pipe := r.client.GetRedisClient().Pipeline()
	for _, key := range k {
		_, _ = pipe.Get(ctx, key).Result()
	}
	cmds, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	out := make(map[string][]byte)
	for _, cmd := range cmds {
		args := cmd.Args()
		if v, ok := args[1].(string); ok {
			err = cmd.Err()
			if errors.Is(err, redis.Nil) {
				continue
			}
			if err != nil {
				return nil, err
			}

			if cmd, ok := cmd.(*redis.StringCmd); ok {
				str := cmd.Val()
				out[v] = []byte(str)
			}
		}
	}
	return out, nil
}
