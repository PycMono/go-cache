package tmpcache

import (
	"context"
	"fmt"
	"github.com/PycMono/go-cache/client"
	"github.com/PycMono/go-cache/client/mem"
	"github.com/PycMono/go-cache/client/redis"
	"github.com/bytedance/sonic"
	"testing"
	"time"
)

func getRedisAdaptor() client.IAdaptor {
	conf := redis.Config{}.WithName("test").
		WithDB(10).
		WithAddr(":6379").
		WithPassword("123456").
		WithPoolSize(100)
	redisClient, err := redis.NewRedisClient(&conf)
	if err != nil {
		panic(err)
	}

	return redis.NewRedisAdaptor(redisClient)
}

func getMemAdaptor() client.IAdaptor {
	conf := mem.Config{}.WithCacheSize(1024 * 1024 * 1024).WithGCPercent(20)

	return mem.NewMemoryAdaptor(mem.NewMemCache(&conf))
}

func TestSet(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	mgr := NewCache[*person](getRedisAdaptor(), &Options{
		Base:      Base{Prefix: "demo"},
		EnableLog: false,
		WriteNil:  false,
		Expire:    time.Minute,
	})

	var (
		m = make(map[string]*person)
	)
	m["12344pyc-test1"] = &person{
		Name: "彭亚川12",
		Age:  20,
	}

	err := mgr.Set(context.TODO(), m)
	if err != nil {
		panic(err)
	}
}

func TestDel(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	mgr := NewCache[*person](getRedisAdaptor(), &Options{
		Base:      Base{Prefix: "demo"},
		EnableLog: false,
		WriteNil:  false,
		Expire:    time.Minute * 10,
	})

	var (
		m = make(map[string]*person)
	)
	m["12344pyc-test1"] = &person{
		Name: "彭亚川12",
		Age:  20,
	}

	err := mgr.Set(context.TODO(), m)
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 30)
	err = mgr.Del(context.TODO(), []string{"12344pyc-test1"})
	if err != nil {
		panic(err)
	}
}

func TestGet(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	mgr := NewCache[*person](getRedisAdaptor(), &Options{
		Base:      Base{Prefix: "demo"},
		EnableLog: false,
		WriteNil:  false,
		Expire:    time.Minute * 10,
	})

	var (
		m = make(map[string]*person)
	)
	m["12344pyc-test1"] = &person{
		Name: "彭亚川12",
		Age:  20,
	}

	err := mgr.Set(context.TODO(), m)
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 10)
	kvMap, err := mgr.Get(context.TODO(), []string{"12344pyc-test1"})
	if err != nil {
		panic(err)
	}

	b, _ := sonic.Marshal(kvMap)
	fmt.Println(string(b))
}

func TestGetAndSetSingle(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	mgr := NewCache[*person](getRedisAdaptor(), &Options{
		Base:      Base{Prefix: "demo"},
		EnableLog: false,
		WriteNil:  true,
		Expire:    time.Minute * 10,
	})

	val, ok, err := mgr.GetAndSetSingle(context.TODO(), "12344pyc-test3", func(k string) (*person, bool, error) {
		return nil, false, nil
	})
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("未找到数据")
	}

	b, _ := sonic.Marshal(val)
	fmt.Println(string(b))
}

func TestName(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	mgr := NewCache[*person](getMemAdaptor(), &Options{
		Base:      Base{Prefix: "demo"},
		EnableLog: false,
		WriteNil:  false,
		Expire:    time.Minute,
	})

	kvMap := make(map[string]*person)
	kvMap["1234"] = &person{
		Name: "113",
		Age:  100,
	}
	err := mgr.Set(context.TODO(), kvMap)
	if err != nil {
		panic(err)
	}

	respMap, _ := mgr.Get(context.TODO(), []string{"1234"})
	b, _ := sonic.Marshal(respMap)
	fmt.Println(string(b))
}
