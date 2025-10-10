package iplocx

import (
	"container/list"
	"sync"
)

// cacheItem 缓存项
type cacheItem struct {
	key   string
	value *Location
}

// LRUCache LRU缓存实现
type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	lruList  *list.List
	mu       sync.RWMutex
	hits     int64 // 缓存命中次数
	misses   int64 // 缓存未命中次数
}

// NewLRUCache 创建LRU缓存
func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		capacity = 1000 // 默认容量
	}
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		lruList:  list.New(),
	}
}

// Get 获取缓存
func (c *LRUCache) Get(key string) (*Location, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[key]; exists {
		c.lruList.MoveToFront(elem)
		c.hits++
		return elem.Value.(*cacheItem).value, true
	}
	c.misses++
	return nil, false
}

// Put 设置缓存
func (c *LRUCache) Put(key string, value *Location) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果已存在，更新并移到前面
	if elem, exists := c.cache[key]; exists {
		c.lruList.MoveToFront(elem)
		elem.Value.(*cacheItem).value = value
		return
	}

	// 新增缓存项
	item := &cacheItem{key: key, value: value}
	elem := c.lruList.PushFront(item)
	c.cache[key] = elem

	// 超出容量，移除最久未使用的
	if c.lruList.Len() > c.capacity {
		oldest := c.lruList.Back()
		if oldest != nil {
			c.lruList.Remove(oldest)
			delete(c.cache, oldest.Value.(*cacheItem).key)
		}
	}
}

// Clear 清空缓存
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*list.Element)
	c.lruList = list.New()
	c.hits = 0
	c.misses = 0
}

// Stats 获取缓存统计信息
func (c *LRUCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total) * 100
	}

	return CacheStats{
		Size:     c.lruList.Len(),
		Capacity: c.capacity,
		Hits:     c.hits,
		Misses:   c.misses,
		HitRate:  hitRate,
	}
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Size     int     // 当前缓存数量
	Capacity int     // 缓存容量
	Hits     int64   // 命中次数
	Misses   int64   // 未命中次数
	HitRate  float64 // 命中率（百分比）
}
