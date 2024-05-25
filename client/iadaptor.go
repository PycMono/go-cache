package client

import (
	"context"
	"time"
)

// IAdaptor 接口转换器
type IAdaptor interface {
	Set(ctx context.Context, params map[string][]byte, expire time.Duration) error
	Get(ctx context.Context, k []string) (map[string][]byte, error)
	Del(ctx context.Context, k []string) error
}
