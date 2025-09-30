package utils

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ShardedRWMutex 分片读写锁 - 减少锁竞争
type ShardedRWMutex struct {
	shards []sync.RWMutex
	mask   uint32
}

// NewShardedRWMutex 创建分片读写锁
func NewShardedRWMutex(shardCount int) *ShardedRWMutex {
	if shardCount <= 0 {
		shardCount = runtime.NumCPU() * 2 // 默认为CPU核心数的2倍
	}
	
	// 确保是2的幂次方
	count := 1
	for count < shardCount {
		count <<= 1
	}
	
	return &ShardedRWMutex{
		shards: make([]sync.RWMutex, count),
		mask:   uint32(count - 1),
	}
}

// hash 计算键的哈希值
func (srm *ShardedRWMutex) hash(key string) uint32 {
	var hash uint32
	for i := 0; i < len(key); i++ {
		hash = hash*31 + uint32(key[i])
	}
	return hash
}

// getLock 获取对应的锁
func (srm *ShardedRWMutex) getLock(key string) *sync.RWMutex {
	index := srm.hash(key) & srm.mask
	return &srm.shards[index]
}

// RLock 读锁
func (srm *ShardedRWMutex) RLock(key string) {
	srm.getLock(key).RLock()
}

// RUnlock 读解锁
func (srm *ShardedRWMutex) RUnlock(key string) {
	srm.getLock(key).RUnlock()
}

// Lock 写锁
func (srm *ShardedRWMutex) Lock(key string) {
	srm.getLock(key).Lock()
}

// Unlock 写解锁
func (srm *ShardedRWMutex) Unlock(key string) {
	srm.getLock(key).Unlock()
}

// OptimizedMutex 优化的互斥锁，带有争用检测
type OptimizedMutex struct {
	mu            sync.Mutex
	contentions   int64
	lockStartTime int64
	maxWaitTime   time.Duration
}

// NewOptimizedMutex 创建优化互斥锁
func NewOptimizedMutex() *OptimizedMutex {
	return &OptimizedMutex{
		maxWaitTime: 10 * time.Millisecond, // 最大等待时间
	}
}

// Lock 加锁
func (om *OptimizedMutex) Lock() {
	start := time.Now()
	om.mu.Lock()
	
	waitTime := time.Since(start)
	if waitTime > om.maxWaitTime {
		atomic.AddInt64(&om.contentions, 1)
	}
	
	atomic.StoreInt64(&om.lockStartTime, start.UnixNano())
}

// Unlock 解锁
func (om *OptimizedMutex) Unlock() {
	om.mu.Unlock()
}

// TryLock 尝试加锁，非阻塞
func (om *OptimizedMutex) TryLock() bool {
	return om.mu.TryLock()
}

// GetContentions 获取争用次数
func (om *OptimizedMutex) GetContentions() int64 {
	return atomic.LoadInt64(&om.contentions)
}

// ResetContentions 重置争用计数
func (om *OptimizedMutex) ResetContentions() {
	atomic.StoreInt64(&om.contentions, 0)
}

// RWMutexWithMetrics 带指标的读写锁
type RWMutexWithMetrics struct {
	mu           sync.RWMutex
	readCount    int64
	writeCount   int64
	readTime     int64  // 总读取时间（纳秒）
	writeTime    int64  // 总写入时间（纳秒）
	currentReads int32  // 当前读取者数量
}

// NewRWMutexWithMetrics 创建带指标的读写锁
func NewRWMutexWithMetrics() *RWMutexWithMetrics {
	return &RWMutexWithMetrics{}
}

// RLock 读锁
func (rwm *RWMutexWithMetrics) RLock() {
	start := time.Now()
	rwm.mu.RLock()
	
	atomic.AddInt64(&rwm.readCount, 1)
	atomic.AddInt64(&rwm.readTime, time.Since(start).Nanoseconds())
	atomic.AddInt32(&rwm.currentReads, 1)
}

// RUnlock 读解锁
func (rwm *RWMutexWithMetrics) RUnlock() {
	rwm.mu.RUnlock()
	atomic.AddInt32(&rwm.currentReads, -1)
}

// Lock 写锁
func (rwm *RWMutexWithMetrics) Lock() {
	start := time.Now()
	rwm.mu.Lock()
	
	atomic.AddInt64(&rwm.writeCount, 1)
	atomic.AddInt64(&rwm.writeTime, time.Since(start).Nanoseconds())
}

// Unlock 写解锁
func (rwm *RWMutexWithMetrics) Unlock() {
	rwm.mu.Unlock()
}

// GetMetrics 获取锁指标
func (rwm *RWMutexWithMetrics) GetMetrics() map[string]interface{} {
	readCount := atomic.LoadInt64(&rwm.readCount)
	writeCount := atomic.LoadInt64(&rwm.writeCount)
	totalReadTime := atomic.LoadInt64(&rwm.readTime)
	totalWriteTime := atomic.LoadInt64(&rwm.writeTime)
	currentReads := atomic.LoadInt32(&rwm.currentReads)
	
	var avgReadTime, avgWriteTime float64
	if readCount > 0 {
		avgReadTime = float64(totalReadTime) / float64(readCount) / 1e6 // 转换为毫秒
	}
	if writeCount > 0 {
		avgWriteTime = float64(totalWriteTime) / float64(writeCount) / 1e6
	}
	
	return map[string]interface{}{
		"read_count":      readCount,
		"write_count":     writeCount,
		"current_reads":   currentReads,
		"avg_read_time":   avgReadTime,
		"avg_write_time":  avgWriteTime,
		"total_read_time": totalReadTime,
		"total_write_time": totalWriteTime,
	}
}

// ResetMetrics 重置指标
func (rwm *RWMutexWithMetrics) ResetMetrics() {
	atomic.StoreInt64(&rwm.readCount, 0)
	atomic.StoreInt64(&rwm.writeCount, 0)
	atomic.StoreInt64(&rwm.readTime, 0)
	atomic.StoreInt64(&rwm.writeTime, 0)
}

// SpinLock 自旋锁 - 适用于短期持有的锁
type SpinLock struct {
	flag int32
}

// NewSpinLock 创建自旋锁
func NewSpinLock() *SpinLock {
	return &SpinLock{}
}

// Lock 加锁
func (sl *SpinLock) Lock() {
	for !atomic.CompareAndSwapInt32(&sl.flag, 0, 1) {
		runtime.Gosched() // 让出CPU
	}
}

// Unlock 解锁
func (sl *SpinLock) Unlock() {
	atomic.StoreInt32(&sl.flag, 0)
}

// TryLock 尝试加锁
func (sl *SpinLock) TryLock() bool {
	return atomic.CompareAndSwapInt32(&sl.flag, 0, 1)
}

// LockManager 锁管理器 - 集中管理不同类型的锁
type LockManager struct {
	// 按功能分类的锁
	petLocks    *ShardedRWMutex  // 宠物数据锁
	eventLocks  *ShardedRWMutex  // 事件数据锁
	cacheLocks  *ShardedRWMutex  // 缓存锁
	globalMutex *RWMutexWithMetrics // 全局锁（带指标）
	
	// 锁使用统计
	lockStats map[string]*LockStats
	statsMux  sync.RWMutex
}

// LockStats 锁统计信息
type LockStats struct {
	LockCount    int64
	UnlockCount  int64
	ContentionCount int64
	TotalLockTime   int64
	MaxLockTime     int64
}

// NewLockManager 创建锁管理器
func NewLockManager() *LockManager {
	return &LockManager{
		petLocks:    NewShardedRWMutex(16),  // 16个分片
		eventLocks:  NewShardedRWMutex(32),  // 32个分片（事件更频繁）
		cacheLocks:  NewShardedRWMutex(8),   // 8个分片
		globalMutex: NewRWMutexWithMetrics(),
		lockStats:   make(map[string]*LockStats),
	}
}

// LockPet 锁定宠物数据
func (lm *LockManager) LockPet(petID string) {
	lm.recordLockStart("pet")
	lm.petLocks.Lock(petID)
}

// UnlockPet 解锁宠物数据
func (lm *LockManager) UnlockPet(petID string) {
	lm.petLocks.Unlock(petID)
	lm.recordLockEnd("pet")
}

// RLockPet 读锁宠物数据
func (lm *LockManager) RLockPet(petID string) {
	lm.recordLockStart("pet_read")
	lm.petLocks.RLock(petID)
}

// RUnlockPet 读解锁宠物数据
func (lm *LockManager) RUnlockPet(petID string) {
	lm.petLocks.RUnlock(petID)
	lm.recordLockEnd("pet_read")
}

// LockEvent 锁定事件数据
func (lm *LockManager) LockEvent(eventKey string) {
	lm.recordLockStart("event")
	lm.eventLocks.Lock(eventKey)
}

// UnlockEvent 解锁事件数据
func (lm *LockManager) UnlockEvent(eventKey string) {
	lm.eventLocks.Unlock(eventKey)
	lm.recordLockEnd("event")
}

// GlobalRLock 全局读锁
func (lm *LockManager) GlobalRLock() {
	lm.globalMutex.RLock()
}

// GlobalRUnlock 全局读解锁
func (lm *LockManager) GlobalRUnlock() {
	lm.globalMutex.RUnlock()
}

// GlobalLock 全局写锁
func (lm *LockManager) GlobalLock() {
	lm.globalMutex.Lock()
}

// GlobalUnlock 全局写解锁
func (lm *LockManager) GlobalUnlock() {
	lm.globalMutex.Unlock()
}

// recordLockStart 记录锁开始
func (lm *LockManager) recordLockStart(lockType string) {
	lm.statsMux.Lock()
	defer lm.statsMux.Unlock()
	
	if stats, exists := lm.lockStats[lockType]; exists {
		atomic.AddInt64(&stats.LockCount, 1)
	} else {
		lm.lockStats[lockType] = &LockStats{
			LockCount: 1,
		}
	}
}

// recordLockEnd 记录锁结束
func (lm *LockManager) recordLockEnd(lockType string) {
	lm.statsMux.Lock()
	defer lm.statsMux.Unlock()
	
	if stats, exists := lm.lockStats[lockType]; exists {
		atomic.AddInt64(&stats.UnlockCount, 1)
	}
}

// GetLockStats 获取锁统计
func (lm *LockManager) GetLockStats() map[string]*LockStats {
	lm.statsMux.RLock()
	defer lm.statsMux.RUnlock()
	
	// 返回副本
	result := make(map[string]*LockStats)
	for k, v := range lm.lockStats {
		statsCopy := *v
		result[k] = &statsCopy
	}
	
	// 添加全局锁指标
	globalMetrics := lm.globalMutex.GetMetrics()
	result["global"] = &LockStats{
		LockCount:   globalMetrics["write_count"].(int64),
		UnlockCount: globalMetrics["read_count"].(int64),
	}
	
	return result
}

// ResetLockStats 重置锁统计
func (lm *LockManager) ResetLockStats() {
	lm.statsMux.Lock()
	defer lm.statsMux.Unlock()
	
	lm.lockStats = make(map[string]*LockStats)
	lm.globalMutex.ResetMetrics()
}

// 全局锁管理器实例
var GlobalLockManager = NewLockManager()