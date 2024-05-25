package tmpcache

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"testing"
	"time"
)

func TestMultiCacheGetAndDel(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	cache := NewMultiCache[*person](&MultiCacheOptions{
		Base:      Base{Prefix: "demo"},
		EnableLog: false,
		WriteNil:  false,
		Expire:    time.Minute * 10,
	}, getMemAdaptor(), getRedisAdaptor())

	var (
		m = make(map[string]*person)
	)
	m["12344pyc-test1"] = &person{
		Name: "彭亚川12",
		Age:  20,
	}
	err := cache.Set(context.TODO(), m)
	if err != nil {
		panic(err)
	}

	kvMap, err := cache.Get(context.TODO(), []string{"12344pyc-test1"})
	if err != nil {
		panic(err)
	}
	b, _ := sonic.Marshal(kvMap)
	fmt.Println(string(b))

	time.Sleep(time.Second * 10)
	err = cache.Del(context.TODO(), []string{"12344pyc-test1"})
	if err != nil {
		panic(err)
	}

	kvMap, err = cache.Get(context.TODO(), []string{"12344pyc-test1"})
	if err != nil {
		panic(err)
	}
	b, _ = sonic.Marshal(kvMap)
	fmt.Println(string(b))
}

func TestMultiCacheGetAndSet(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	cache := NewMultiCache[*person](&MultiCacheOptions{
		Base:      Base{Prefix: "demo"},
		EnableLog: false,
		WriteNil:  false,
		Expire:    time.Minute * 10,
	}, getMemAdaptor(), getRedisAdaptor())

	kvMap, err := cache.GetAndSet(context.TODO(), []string{"12344pyc-test1"}, func(k []string) (map[string]*person, error) {
		var (
			m = make(map[string]*person)
		)
		m["12344pyc-test1"] = &person{
			Name: "彭亚川12",
			Age:  20,
		}
		return m, nil
	})
	if err != nil {
		panic(err)
	}
	b, _ := sonic.Marshal(kvMap)
	fmt.Println(string(b))
	fmt.Println("------------1")

	kvMap, err = cache.Get(context.TODO(), []string{"12344pyc-test1"})
	if err != nil {
		panic(err)
	}
	b, _ = sonic.Marshal(kvMap)
	fmt.Println(string(b))
	fmt.Println("------------2")
}
