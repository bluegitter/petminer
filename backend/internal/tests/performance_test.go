package tests

import (
	"fmt"
	"log"
	"runtime"
	"testing"
	"time"

	"miningpet/internal/cache"
	"miningpet/internal/database"
	"miningpet/internal/models"
	"miningpet/internal/services"
	"miningpet/internal/utils"
)

// BenchmarkPetCreation 测试宠物创建性能
func BenchmarkPetCreation(b *testing.B) {
	// 初始化数据库
	if err := database.Initialize(); err != nil {
		b.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	
	if err := database.Migrate(); err != nil {
		b.Fatalf("Failed to migrate database: %v", err)
	}
	
	petService := services.NewPetService()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		ownerName := fmt.Sprintf("owner_%d", i)
		_, err := petService.CreatePet(ownerName)
		if err != nil {
			b.Errorf("Failed to create pet: %v", err)
		}
	}
}

// BenchmarkGetPet 测试获取宠物性能
func BenchmarkGetPet(b *testing.B) {
	// 初始化数据库
	if err := database.Initialize(); err != nil {
		b.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	
	if err := database.Migrate(); err != nil {
		b.Fatalf("Failed to migrate database: %v", err)
	}
	
	petService := services.NewPetService()
	
	// 预创建一些宠物
	petIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		pet, err := petService.CreatePet(fmt.Sprintf("owner_%d", i))
		if err != nil {
			b.Fatalf("Failed to create pet: %v", err)
		}
		petIDs[i] = pet.ID
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		petID := petIDs[i%100]
		_, exists := petService.GetPet(petID)
		if !exists {
			b.Errorf("Pet not found: %s", petID)
		}
	}
}

// BenchmarkMemoryCache 测试内存缓存性能
func BenchmarkMemoryCache(b *testing.B) {
	cache := cache.NewMemoryCache(5*time.Minute, 1*time.Minute)
	defer cache.Stop()
	
	b.ResetTimer()
	
	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i)
			value := fmt.Sprintf("value_%d", i)
			cache.Set(key, value, 0)
		}
	})
	
	b.Run("Get", func(b *testing.B) {
		// 预设一些数据
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("key_%d", i)
			value := fmt.Sprintf("value_%d", i)
			cache.Set(key, value, 0)
		}
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%1000)
			_, exists := cache.Get(key)
			if !exists {
				b.Errorf("Key not found: %s", key)
			}
		}
	})
}

// BenchmarkObjectPool 测试对象池性能
func BenchmarkObjectPool(b *testing.B) {
	pool := utils.NewObjectPool()
	
	b.Run("WithoutPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			event := &models.Event{
				ID:      fmt.Sprintf("event_%d", i),
				Message: "Test event",
			}
			_ = event
		}
	})
	
	b.Run("WithPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			event := pool.GetEvent()
			event.ID = fmt.Sprintf("event_%d", i)
			event.Message = "Test event"
			pool.PutEvent(event)
		}
	})
}

// BenchmarkJSONSerialization 测试JSON序列化性能
func BenchmarkJSONSerialization(b *testing.B) {
	optimizer := utils.NewJSONOptimizer()
	
	event := &models.Event{
		ID:      "test_event",
		PetID:   "test_pet",
		PetName: "Test Pet",
		Type:    models.EventBattle,
		Message: "Test battle event",
	}
	
	b.Run("StandardJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := utils.FastMarshalEvent(event)
			if err != nil {
				b.Errorf("JSON marshal failed: %v", err)
			}
		}
	})
	
	b.Run("OptimizedJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := optimizer.FastMarshalEvent(event)
			if err != nil {
				b.Errorf("Optimized JSON marshal failed: %v", err)
			}
		}
	})
}

// BenchmarkLockContention 测试锁竞争性能
func BenchmarkLockContention(b *testing.B) {
	b.Run("StandardMutex", func(b *testing.B) {
		var counter int64
		mutex := utils.NewOptimizedMutex()
		
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mutex.Lock()
				counter++
				mutex.Unlock()
			}
		})
	})
	
	b.Run("ShardedMutex", func(b *testing.B) {
		shardedMutex := utils.NewShardedRWMutex(16)
		counters := make(map[string]int64)
		
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key_%d", i%16)
				shardedMutex.Lock(key)
				counters[key]++
				shardedMutex.Unlock(key)
				i++
			}
		})
	})
}

// TestPerformanceReport 生成性能报告
func TestPerformanceReport(t *testing.T) {
	log.Println("=== 性能优化报告 ===")
	
	// 内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	log.Printf("内存使用情况:")
	log.Printf("  分配的堆内存: %d KB", bToKB(m.Alloc))
	log.Printf("  累计分配内存: %d KB", bToKB(m.TotalAlloc))
	log.Printf("  系统内存: %d KB", bToKB(m.Sys))
	log.Printf("  GC次数: %d", m.NumGC)
	
	// 测试各组件
	testCachePerformance(t)
	testLockPerformance(t)
	testStorageStrategy(t)
}

func testCachePerformance(t *testing.T) {
	log.Printf("\n缓存性能测试:")
	
	cacheManager := cache.NewGameCacheManager()
	defer cacheManager.Stop()
	
	// 测试缓存写入
	start := time.Now()
	for i := 0; i < 1000; i++ {
		pet := &models.Pet{
			ID:   fmt.Sprintf("pet_%d", i),
			Name: fmt.Sprintf("Pet %d", i),
		}
		cacheManager.SetPet(pet.ID, pet)
	}
	writeTime := time.Since(start)
	
	// 测试缓存读取
	start = time.Now()
	hitCount := 0
	for i := 0; i < 1000; i++ {
		petID := fmt.Sprintf("pet_%d", i)
		if _, exists := cacheManager.GetPet(petID); exists {
			hitCount++
		}
	}
	readTime := time.Since(start)
	
	log.Printf("  写入1000个对象耗时: %v", writeTime)
	log.Printf("  读取1000个对象耗时: %v", readTime)
	log.Printf("  缓存命中率: %.2f%%", float64(hitCount)/10.0)
	
	// 获取缓存统计
	stats := cacheManager.GetCacheStats()
	log.Printf("  缓存统计: %+v", stats)
}

func testLockPerformance(t *testing.T) {
	log.Printf("\n锁性能测试:")
	
	lockManager := utils.NewLockManager()
	
	// 测试锁操作
	start := time.Now()
	for i := 0; i < 1000; i++ {
		petID := fmt.Sprintf("pet_%d", i%10) // 10个不同的pet ID
		lockManager.LockPet(petID)
		// 模拟一些工作
		time.Sleep(100 * time.Microsecond)
		lockManager.UnlockPet(petID)
	}
	lockTime := time.Since(start)
	
	log.Printf("  1000次锁操作耗时: %v", lockTime)
	
	// 获取锁统计
	lockStats := lockManager.GetLockStats()
	for lockType, stats := range lockStats {
		log.Printf("  %s锁统计: LockCount=%d, UnlockCount=%d", 
			lockType, stats.LockCount, stats.UnlockCount)
	}
}

func testStorageStrategy(t *testing.T) {
	log.Printf("\n存储策略测试:")
	
	strategyManager := cache.NewStrategyManager()
	defer strategyManager.Stop()
	
	// 测试不同优先级的数据处理
	start := time.Now()
	
	// 关键数据
	for i := 0; i < 100; i++ {
		data := fmt.Sprintf("critical_data_%d", i)
		strategyManager.AddData(data, cache.Critical)
	}
	
	// 重要数据
	for i := 0; i < 500; i++ {
		data := fmt.Sprintf("important_data_%d", i)
		strategyManager.AddData(data, cache.Important)
	}
	
	// 瞬时数据
	for i := 0; i < 1000; i++ {
		data := fmt.Sprintf("transient_data_%d", i)
		strategyManager.AddData(data, cache.Transient)
	}
	
	addTime := time.Since(start)
	
	// 等待处理
	time.Sleep(2 * time.Second)
	
	log.Printf("  添加1600个不同优先级数据耗时: %v", addTime)
	
	// 获取存储统计
	storageStats := strategyManager.GetStats()
	for priority, stats := range storageStats {
		log.Printf("  优先级%d统计: 处理数量=%d, 错误数量=%d, 队列长度=%d", 
			int(priority), stats.ProcessedCount, stats.ErrorCount, stats.QueueLength)
	}
}

func bToKB(b uint64) uint64 {
	return b / 1024
}

// TestIntegrationPerformance 集成性能测试
func TestIntegrationPerformance(t *testing.T) {
	// 初始化数据库
	if err := database.Initialize(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	
	if err := database.Migrate(); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}
	
	log.Println("\n=== 集成性能测试 ===")
	
	petService := services.NewPetService()
	
	// 测试高并发场景
	start := time.Now()
	
	// 创建宠物
	petCount := 50
	for i := 0; i < petCount; i++ {
		_, err := petService.CreatePet(fmt.Sprintf("owner_%d", i))
		if err != nil {
			t.Errorf("Failed to create pet: %v", err)
		}
	}
	
	createTime := time.Since(start)
	
	// 获取系统统计
	stats := petService.GetSystemStats()
	
	log.Printf("创建%d个宠物耗时: %v", petCount, createTime)
	log.Printf("系统统计信息:")
	log.Printf("  宠物数据: %+v", stats["pets"])
	log.Printf("  事件数据: %+v", stats["events"]) 
	log.Printf("  性能特性: %+v", stats["performance"])
	
	if lockStats, ok := stats["locks"].(map[string]interface{}); ok {
		log.Printf("  锁统计: %+v", lockStats)
	}
}