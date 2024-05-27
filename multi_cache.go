package tmpcache

import (
	"context"
	"fmt"
	"github.com/PycMono/go-cache/client"
	"github.com/bytedance/sonic"
	"golang.org/x/sync/singleflight"
	"time"
)

// MultiCacheOptions 可选参数
type MultiCacheOptions struct {
	Base
	EnableLog bool          // 是否输出日志
	WriteNil  bool          // 缓存miss是否写入nil防止缓存穿透，默认不写入
	Expire    time.Duration // 过期时间
}

// MultiCache 多级缓存
type MultiCache[T any] struct {
	// 多级缓存适配器
	// 后续处理逻辑是根据数组的顺序遍历，建议把离用户最近的缓存设置到下标0的位置，依次排列。需注意，一定要保证离用户最近的缓存有数据
	// 假设缓存顺序，内存缓存、redis缓存、缓存 miss 透传数据库
	// 若内存缓存未miss，直接返回
	// 若内存缓存miss，从redis中查询，redis 缓存miss 再从数据库中查询，一定要回写内存缓存
	handlers []client.IAdaptor
	opts     *MultiCacheOptions // 基础配置
	sf       singleflight.Group
	// 还可以增加一些中间件比如日志输出、埋点之类的
}

func NewMultiCache[T any](opts *MultiCacheOptions, handlers ...client.IAdaptor) ICache[T] {
	return &MultiCache[T]{
		handlers: handlers,
		opts:     opts,
	}
}

func (c *MultiCache[T]) Set(ctx context.Context, params map[string]T) error {
	kv := make(map[string][]byte)
	for k, v := range params {
		key := c.opts.buildKey(k)
		b, err := sonic.Marshal(v)
		if err != nil {
			return err
		}
		kv[key] = b
	}

	for _, v := range c.handlers {
		err := v.Set(ctx, kv, c.opts.Expire)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *MultiCache[T]) Get(ctx context.Context, keys []string) (map[string]T, error) {
	var (
		tmpKeys = c.opts.buildKeys(keys)
	)

	// 多级缓存查询
	// 思路：第一个client先查找，若miss，将miss的key集合投递下一个client查找，直到所有client查找完成，或者keys全部找到
	var (
		kvMap     = make(map[string][]byte)
		missKeys  = tmpKeys
		preClient client.IAdaptor
	)
	for _, cli := range c.handlers {
		if len(missKeys) == 0 {
			break // 退出循环
		}

		tmpKvMap, err := cli.Get(ctx, missKeys)
		if err != nil {
			return nil, err
		}
		for k, v := range tmpKvMap {
			kvMap[k] = v
		}

		missKeys = []string{} // 重新设置值
		for _, key := range tmpKeys {
			if _, ok := tmpKvMap[key]; !ok {
				missKeys = append(missKeys, key)
			}
		}

		if len(tmpKvMap) == 0 || preClient == nil {
			preClient = cli
			continue
		}

		err = preClient.Set(ctx, tmpKvMap, c.opts.Expire) // 如果1级缓存miss了，2级缓存加载后，回写1级缓存
		if err != nil {
			fmt.Println(err)
		}
		preClient = cli
	}

	// 处理数据返回
	out := make(map[string]T)
	for k, v := range kvMap {
		var obj T
		err := sonic.Unmarshal(v, &obj)
		if err != nil {
			return nil, err
		}

		key := c.opts.splitKey(k) // 切割key
		out[key] = obj
	}

	return out, nil
}

// GetAndSet 缓存 miss，支持调用f函数从其它db中获取数据
func (c *MultiCache[T]) GetAndSet(ctx context.Context, k []string, f func(k []string) (map[string]T, error)) (map[string]T, error) {
	kvMap, err := c.Get(ctx, k)
	if err != nil {
		return nil, err
	}

	var (
		missKeys = []string{}
	)
	for _, v := range k {
		if _, ok := kvMap[v]; ok {
			continue
		}
		missKeys = append(missKeys, v)
	}
	if len(missKeys) == 0 {
		return kvMap, nil
	}

	tmpKvMap, err := f(missKeys)
	if err != nil {
		return nil, err
	}
	for k, v := range tmpKvMap {
		kvMap[k] = v
	}

	// 检查外部数据源数据查询是否一致
	if c.opts.WriteNil && len(missKeys) != len(tmpKvMap) {
		for _, v := range missKeys {
			if _, ok := tmpKvMap[v]; ok {
				continue
			}
			var obj T
			tmpKvMap[v] = obj
		}
	}
	if len(tmpKvMap) > 0 {
		err = c.Set(ctx, tmpKvMap)
		if err != nil {
			// todo 打印日志就好了，不影响后续流程，下次请求再次尝试加载到缓存
			fmt.Println(err)
		}
	}

	return kvMap, nil
}

func (c *MultiCache[T]) GetAndSetSingle(ctx context.Context, k string, f func(k string) (T, bool, error)) (T, bool, error) {
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

func (c *MultiCache[T]) Del(ctx context.Context, k []string) error {
	var (
		tmpKeys = c.opts.buildKeys(k)
	)
	for _, v := range c.handlers {
		err := v.Del(ctx, tmpKeys)
		if err != nil {
			return err
		}
	}

	return nil
}
