package cache

import (
	"sync"
	"time"
)

// CacheItem 缓存项
type CacheItem struct {
	Value      interface{}
	Expiration int64
}

// IsExpired 检查是否过期
func (item CacheItem) IsExpired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// MemoryCache 内存缓存
type MemoryCache struct {
	mu          sync.RWMutex
	items       map[string]CacheItem
	defaultTTL  time.Duration
	cleanupStop chan bool
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache(defaultTTL time.Duration, cleanupInterval time.Duration) *MemoryCache {
	cache := &MemoryCache{
		items:       make(map[string]CacheItem),
		defaultTTL:  defaultTTL,
		cleanupStop: make(chan bool),
	}
	
	// 启动清理协程
	go cache.startCleanup(cleanupInterval)
	
	return cache
}

// Set 设置缓存项
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	} else if c.defaultTTL > 0 {
		expiration = time.Now().Add(c.defaultTTL).UnixNano()
	}
	
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

// Get 获取缓存项
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, exists := c.items[key]
	if !exists {
		return nil, false
	}
	
	if item.IsExpired() {
		// 延迟删除过期项
		go func() {
			c.mu.Lock()
			delete(c.items, key)
			c.mu.Unlock()
		}()
		return nil, false
	}
	
	return item.Value, true
}

// Delete 删除缓存项
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear 清空缓存
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]CacheItem)
}

// Size 获取缓存大小
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Keys 获取所有键
func (c *MemoryCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

// Stop 停止缓存（停止清理协程）
func (c *MemoryCache) Stop() {
	close(c.cleanupStop)
}

// startCleanup 启动清理协程
func (c *MemoryCache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.cleanupStop:
			return
		}
	}
}

// cleanup 清理过期项
func (c *MemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for key, item := range c.items {
		if item.IsExpired() {
			delete(c.items, key)
		}
	}
}

// Stats 缓存统计信息
type CacheStats struct {
	ItemCount   int
	HitCount    int64
	MissCount   int64
	HitRate     float64
}

// StatsMemoryCache 带统计的内存缓存
type StatsMemoryCache struct {
	*MemoryCache
	hitCount  int64
	missCount int64
	mu        sync.RWMutex
}

// NewStatsMemoryCache 创建带统计的内存缓存
func NewStatsMemoryCache(defaultTTL time.Duration, cleanupInterval time.Duration) *StatsMemoryCache {
	return &StatsMemoryCache{
		MemoryCache: NewMemoryCache(defaultTTL, cleanupInterval),
	}
}

// Get 获取缓存项（带统计）
func (sc *StatsMemoryCache) Get(key string) (interface{}, bool) {
	value, exists := sc.MemoryCache.Get(key)
	
	sc.mu.Lock()
	if exists {
		sc.hitCount++
	} else {
		sc.missCount++
	}
	sc.mu.Unlock()
	
	return value, exists
}

// GetStats 获取缓存统计
func (sc *StatsMemoryCache) GetStats() CacheStats {
	sc.mu.RLock()
	hitCount := sc.hitCount
	missCount := sc.missCount
	sc.mu.RUnlock()
	
	totalCount := hitCount + missCount
	hitRate := 0.0
	if totalCount > 0 {
		hitRate = float64(hitCount) / float64(totalCount)
	}
	
	return CacheStats{
		ItemCount: sc.Size(),
		HitCount:  hitCount,
		MissCount: missCount,
		HitRate:   hitRate,
	}
}

// ResetStats 重置统计
func (sc *StatsMemoryCache) ResetStats() {
	sc.mu.Lock()
	sc.hitCount = 0
	sc.missCount = 0
	sc.mu.Unlock()
}