package mem

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestMem(t *testing.T) {
	memoryAdaptor := NewMemoryAdaptor(NewMemCache(&Config{
		cacheSize: 1024 * 1024 * 1024,
		gcPercent: 40,
	}))

	kvMap := make(map[string][]byte)
	kvMap["123"] = []byte("3333")
	memoryAdaptor.Set(context.TODO(), kvMap, time.Hour)
	fmt.Println("-------------------")
	time.Sleep(time.Second * 10)
	respMap, _ := memoryAdaptor.Get(context.TODO(), []string{"123"})
	b, _ := json.Marshal(respMap)
	fmt.Println(string(b))
	fmt.Println("-------------------")

	time.Sleep(time.Second * 10)
	memoryAdaptor.Del(context.TODO(), []string{"123"})

	fmt.Println("-------------------")
	respMap, _ = memoryAdaptor.Get(context.TODO(), []string{"123"})
	b, _ = json.Marshal(respMap)
	fmt.Println(string(b))
	fmt.Println("-------------------")
}
