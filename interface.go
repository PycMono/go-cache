package tmpcache

import "context"

// ICache 管理器接口
type ICache[T any] interface {
	Set(ctx context.Context, params map[string]T) error
	Get(ctx context.Context, keys []string) (map[string]T, error)
	GetAndSet(ctx context.Context, keys []string, f func(keys []string) (map[string]T, error)) (map[string]T, error)
	GetAndSetSingle(ctx context.Context, k string, f func(k string) (T, bool, error)) (T, bool, error)
	Del(ctx context.Context, keys []string) error
}
