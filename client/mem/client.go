package mem

import (
	"github.com/coocood/freecache"
	"runtime/debug"
)

type Client struct {
	conf        *Config
	cacheClient *freecache.Cache
}

func NewMemCache(conf *Config) *Client {
	cacheClient := freecache.NewCache(conf.cacheSize)
	debug.SetGCPercent(conf.gcPercent)

	return &Client{
		conf:        conf,
		cacheClient: cacheClient,
	}
}
