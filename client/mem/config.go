package mem

// Config 配置文件
type Config struct {
	cacheSize int // 缓存块大小
	gcPercent int // 垃圾收集目标百分比
}

func (c Config) WithCacheSize(cacheSize int) Config {
	c.cacheSize = cacheSize
	return c
}

func (c Config) WithGCPercent(gcPercent int) Config {
	c.gcPercent = gcPercent
	return c
}
