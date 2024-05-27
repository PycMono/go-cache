package mem

import (
	"context"
	"errors"
	"github.com/PycMono/go-cache/client"
	"github.com/coocood/freecache"
	"time"
)

type Cache struct {
	client *Client
}

func NewMemoryAdaptor(client *Client) client.IAdaptor {
	return &Cache{client: client}
}

func (r *Cache) Set(ctx context.Context, params map[string][]byte, expire time.Duration) error {
	for k, v := range params {
		err := r.client.cacheClient.Set([]byte(k), v, int(expire.Seconds()))
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Cache) Del(ctx context.Context, k []string) error {
	for _, v := range k {
		r.client.cacheClient.Del([]byte(v))
	}

	return nil
}

func (r *Cache) Get(ctx context.Context, k []string) (map[string][]byte, error) {
	out := make(map[string][]byte)
	for _, id := range k {
		val, err := r.client.cacheClient.Get([]byte(id))
		if err != nil {
			if errors.As(err, &freecache.ErrNotFound) {
				continue
			}
			return nil, err
		}
		out[id] = val
	}

	return out, nil
}
