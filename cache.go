package tmpcache

import (
	"PycMono/github/go/go-cache/client"
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"golang.org/x/sync/singleflight"
	"strings"
	"time"
)

// Options 配置文件
type Options struct {
	Base
	EnableLog bool          // 是否输出日志
	WriteNil  bool          // 缓存miss是否写入nil防止缓存穿透，默认不写入
	Expire    time.Duration // 过期时间
}

type Base struct {
	Prefix string // 缓存前缀，支持自定义，如果带了前缀会拼接在key上
}

// 构造key
func (c *Base) buildKeys(keys []string) []string {
	var (
		tmpKeys = []string{}
	)
	for _, v := range keys {
		tmpKeys = append(tmpKeys, c.buildKey(v))
	}
	return tmpKeys
}

// 构造key
func (c *Base) buildKey(k string) string {
	if len(c.Prefix) > 0 {
		return fmt.Sprintf("%s_%s", c.Prefix, k)
	}
	return k
}

func (c *Base) splitKey(k string) string {
	sts := strings.Split(k, "_")
	if len(sts) > 0 {
		return sts[1]
	}
	return sts[0]
}

type Cache[T any] struct {
	handler client.IAdaptor // 适配器client
	opts    *Options        // 基础配置
	sf      singleflight.Group
	// 还可以增加一些中间件比如日志、埋点之类的
}

func NewCache[T any](handler client.IAdaptor, opts *Options) ICache[T] {
	return &Cache[T]{
		handler: handler,
		opts:    opts,
	}
}

func (c *Cache[T]) Set(ctx context.Context, params map[string]T) error {
	kv := make(map[string][]byte)
	for k, v := range params {
		key := c.opts.buildKey(k)
		b, err := sonic.Marshal(v)
		if err != nil {
			return err
		}
		kv[key] = b
	}

	return c.handler.Set(ctx, kv, c.opts.Expire)
}

func (c *Cache[T]) Get(ctx context.Context, keys []string) (map[string]T, error) {
	var (
		tmpKeys = c.opts.buildKeys(keys)
	)
	kv, err := c.handler.Get(ctx, tmpKeys)
	if err != nil {
		return nil, err
	}

	out := make(map[string]T)
	for k, v := range kv {
		var obj T
		err = sonic.Unmarshal(v, &obj)
		if err != nil {
			return nil, err
		}

		key := c.opts.splitKey(k) // 切割key
		out[key] = obj
	}

	return out, nil
}

// GetAndSet 缓存 miss，支持调用f函数从其它db中获取数据
func (c *Cache[T]) GetAndSet(ctx context.Context, keys []string, f func(keys []string) (map[string]T, error)) (map[string]T, error) {
	kv, err := c.Get(ctx, keys)
	if err != nil {
		return nil, err
	}

	var (
		missKeys = []string{}
	)
	for _, v := range keys {
		if _, ok := kv[v]; ok {
			continue
		}
		missKeys = append(missKeys, v)
	}
	if len(missKeys) == 0 {
		return kv, nil
	}

	tmpKv, err := f(missKeys)
	if err != nil {
		return nil, err
	}
	for k, v := range tmpKv {
		kv[k] = v
	}

	// 检查外部数据源数据查询是否一致
	if c.opts.WriteNil && len(missKeys) != len(tmpKv) {
		for _, v := range missKeys {
			if _, ok := tmpKv[v]; ok {
				continue
			}
			var obj T
			tmpKv[v] = obj
		}
	}
	if len(tmpKv) > 0 {
		err = c.Set(ctx, tmpKv)
		if err != nil {
			// todo 打印日志就好了，不影响后续流程，下次请求再次尝试加载到缓存
			fmt.Println(err)
		}
	}

	return kv, nil
}

func (c *Cache[T]) GetAndSetSingle(ctx context.Context, k string, f func(k string) (T, bool, error)) (T, bool, error) {
	var (
		val T
		ok  bool
	)

	kvMap, err := c.Get(ctx, []string{k})
	if err != nil {
		return val, false, err
	}
	val, ok = kvMap[k]
	if ok {
		return val, true, nil
	}

	// 缓存miss 从外部查询
	if f == nil {
		return val, false, nil
	}

	// 单飞查询
	_, err, _ = c.sf.Do(k, func() (interface{}, error) {
		val, ok, err = f(k)
		if err != nil {
			return nil, err
		}
		return val, nil
	})
	if err != nil {
		return val, false, err
	}

	// 写入缓存
	var (
		tmpKvMap = make(map[string]T)
	)

	// 写入缓存条件：1、数据存在；2、数据不存在并且WriteNil为true
	if ok || (c.opts.WriteNil && !ok) {
		tmpKvMap[k] = val
		err = c.Set(ctx, tmpKvMap)
		if err != nil {
			fmt.Println(err)
		}
	}

	return val, ok, nil
}

func (c *Cache[T]) Del(ctx context.Context, keys []string) error {
	var (
		tmpKeys = c.opts.buildKeys(keys)
	)
	return c.handler.Del(ctx, tmpKeys)
}
