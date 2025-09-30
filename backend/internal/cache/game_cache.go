package cache

import (
	"fmt"
	"log"
	"miningpet/internal/models"
	"time"
)

// GameCacheManager 游戏缓存管理器
type GameCacheManager struct {
	// 宠物缓存 - 长期缓存，因为宠物数据相对稳定
	petCache *StatsMemoryCache
	
	// 事件缓存 - 短期缓存，用于快速访问最近事件
	eventCache *StatsMemoryCache
	
	// 玩家状态缓存 - 超短期缓存，用于高频读写的状态数据
	playerStateCache *StatsMemoryCache
	
	// 游戏统计缓存 - 中期缓存，用于统计数据
	statsCache *StatsMemoryCache
}

// NewGameCacheManager 创建游戏缓存管理器
func NewGameCacheManager() *GameCacheManager {
	return &GameCacheManager{
		// 宠物缓存：30分钟过期，每10分钟清理
		petCache: NewStatsMemoryCache(30*time.Minute, 10*time.Minute),
		
		// 事件缓存：5分钟过期，每分钟清理
		eventCache: NewStatsMemoryCache(5*time.Minute, 1*time.Minute),
		
		// 玩家状态缓存：2分钟过期，每30秒清理（高频访问）
		playerStateCache: NewStatsMemoryCache(2*time.Minute, 30*time.Second),
		
		// 统计缓存：15分钟过期，每5分钟清理
		statsCache: NewStatsMemoryCache(15*time.Minute, 5*time.Minute),
	}
}

// 宠物缓存相关方法
func (gcm *GameCacheManager) SetPet(petID string, pet *models.Pet) {
	gcm.petCache.Set(fmt.Sprintf("pet:%s", petID), pet, 0)
}

func (gcm *GameCacheManager) GetPet(petID string) (*models.Pet, bool) {
	if value, exists := gcm.petCache.Get(fmt.Sprintf("pet:%s", petID)); exists {
		if pet, ok := value.(*models.Pet); ok {
			return pet, true
		}
	}
	return nil, false
}

func (gcm *GameCacheManager) DeletePet(petID string) {
	gcm.petCache.Delete(fmt.Sprintf("pet:%s", petID))
}

// 根据Owner缓存宠物
func (gcm *GameCacheManager) SetPetByOwner(owner string, pet *models.Pet) {
	gcm.petCache.Set(fmt.Sprintf("pet:owner:%s", owner), pet, 0)
}

func (gcm *GameCacheManager) GetPetByOwner(owner string) (*models.Pet, bool) {
	if value, exists := gcm.petCache.Get(fmt.Sprintf("pet:owner:%s", owner)); exists {
		if pet, ok := value.(*models.Pet); ok {
			return pet, true
		}
	}
	return nil, false
}

// 事件缓存相关方法
func (gcm *GameCacheManager) SetRecentEvents(events []models.Event) {
	gcm.eventCache.Set("recent_events", events, 0)
}

func (gcm *GameCacheManager) GetRecentEvents() ([]models.Event, bool) {
	if value, exists := gcm.eventCache.Get("recent_events"); exists {
		if events, ok := value.([]models.Event); ok {
			return events, true
		}
	}
	return nil, false
}

func (gcm *GameCacheManager) SetPetEvents(petID string, events []*models.Event) {
	gcm.eventCache.Set(fmt.Sprintf("events:pet:%s", petID), events, 0)
}

func (gcm *GameCacheManager) GetPetEvents(petID string) ([]*models.Event, bool) {
	if value, exists := gcm.eventCache.Get(fmt.Sprintf("events:pet:%s", petID)); exists {
		if events, ok := value.([]*models.Event); ok {
			return events, true
		}
	}
	return nil, false
}

// 玩家状态缓存相关方法
func (gcm *GameCacheManager) SetPlayerState(petID string, state interface{}) {
	gcm.playerStateCache.Set(fmt.Sprintf("state:%s", petID), state, 0)
}

func (gcm *GameCacheManager) GetPlayerState(petID string) (interface{}, bool) {
	return gcm.playerStateCache.Get(fmt.Sprintf("state:%s", petID))
}

// 统计缓存相关方法
func (gcm *GameCacheManager) SetGameStats(key string, stats interface{}) {
	gcm.statsCache.Set(fmt.Sprintf("stats:%s", key), stats, 0)
}

func (gcm *GameCacheManager) GetGameStats(key string) (interface{}, bool) {
	return gcm.statsCache.Get(fmt.Sprintf("stats:%s", key))
}

// 缓存统计和管理
func (gcm *GameCacheManager) GetCacheStats() map[string]CacheStats {
	return map[string]CacheStats{
		"pets":         gcm.petCache.GetStats(),
		"events":       gcm.eventCache.GetStats(),
		"playerState":  gcm.playerStateCache.GetStats(),
		"stats":        gcm.statsCache.GetStats(),
	}
}

func (gcm *GameCacheManager) LogCacheStats() {
	stats := gcm.GetCacheStats()
	log.Printf("Cache Statistics:")
	for name, stat := range stats {
		log.Printf("  %s: Items=%d, Hits=%d, Misses=%d, HitRate=%.2f%%", 
			name, stat.ItemCount, stat.HitCount, stat.MissCount, stat.HitRate*100)
	}
}

func (gcm *GameCacheManager) ClearAllCache() {
	gcm.petCache.Clear()
	gcm.eventCache.Clear()
	gcm.playerStateCache.Clear()
	gcm.statsCache.Clear()
	log.Println("All caches cleared")
}

func (gcm *GameCacheManager) Stop() {
	gcm.petCache.Stop()
	gcm.eventCache.Stop()
	gcm.playerStateCache.Stop()
	gcm.statsCache.Stop()
	log.Println("Cache managers stopped")
}

// 预热缓存接口
type CacheWarmer interface {
	WarmupCache(gcm *GameCacheManager) error
}

// 缓存预热函数
func (gcm *GameCacheManager) WarmupWithData(pets []*models.Pet, recentEvents []models.Event) {
	log.Println("Starting cache warmup...")
	
	// 预热宠物缓存
	for _, pet := range pets {
		gcm.SetPet(pet.ID, pet)
		gcm.SetPetByOwner(pet.Owner, pet)
	}
	
	// 预热事件缓存
	if len(recentEvents) > 0 {
		gcm.SetRecentEvents(recentEvents)
	}
	
	log.Printf("Cache warmup completed: %d pets, %d events", len(pets), len(recentEvents))
}