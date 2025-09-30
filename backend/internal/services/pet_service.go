package services

import (
	"fmt"
	"log"
	"time"

	"miningpet/internal/cache"
	"miningpet/internal/database"
	"miningpet/internal/models"
	"miningpet/internal/utils"
	"github.com/google/uuid"
)

type PetService struct {
	pets         map[string]*models.Pet
	events       []models.Event
	mutex        *utils.RWMutexWithMetrics // 使用优化的锁
	eventsCh     chan models.Event
	aiEngine     *AIEngine
	activePets   map[string]*time.Ticker
	recentEvents map[string]time.Time
	
	petRepo   *database.PetRepository
	eventRepo *database.EventRepository
	
	// 内存缓存管理器
	cacheManager *cache.GameCacheManager
	// 状态管理器
	stateManager *cache.StateManager
	// 存储策略管理器
	strategyManager *cache.StrategyManager
	// 对象池
	objectPool *utils.ObjectPool
	// JSON优化器
	jsonOptimizer *utils.JSONOptimizer
}

func NewPetService() *PetService {
	ps := &PetService{
		pets:            make(map[string]*models.Pet),
		events:          make([]models.Event, 0),
		mutex:           utils.NewRWMutexWithMetrics(),
		eventsCh:        make(chan models.Event, 100),
		aiEngine:        NewAIEngine(),
		activePets:      make(map[string]*time.Ticker),
		recentEvents:    make(map[string]time.Time),
		petRepo:         database.NewPetRepository(),
		eventRepo:       database.NewEventRepository(),
		cacheManager:    cache.NewGameCacheManager(),
		stateManager:    cache.NewStateManager(),
		strategyManager: cache.NewStrategyManager(),
		objectPool:      utils.NewObjectPool(),
		jsonOptimizer:   utils.NewJSONOptimizer(),
	}
	
	if err := ps.loadPetsFromDatabase(); err != nil {
		log.Printf("Warning: failed to load pets from database: %v", err)
	}
	
	// 预热缓存
	ps.warmupCache()
	
	go ps.runGlobalAI()
	ps.startExistingPetsAI()
	
	return ps
}

func (ps *PetService) loadPetsFromDatabase() error {
	pets, err := ps.petRepo.GetAllPets()
	if err != nil {
		return fmt.Errorf("failed to load pets from database: %w", err)
	}

	ps.mutex.Lock()
	for _, pet := range pets {
		ps.pets[pet.ID] = pet
		// 同时缓存到内存
		ps.cacheManager.SetPet(pet.ID, pet)
		ps.cacheManager.SetPetByOwner(pet.Owner, pet)
	}
	ps.mutex.Unlock()

	log.Printf("Loaded %d pets from database", len(pets))
	return nil
}

// warmupCache 预热缓存
func (ps *PetService) warmupCache() {
	// 获取最近事件并缓存
	recentEvents := ps.GetRecentEvents(100)
	if len(recentEvents) > 0 {
		ps.cacheManager.SetRecentEvents(recentEvents)
	}
	
	log.Printf("Cache warmed up with %d recent events", len(recentEvents))
}

func (ps *PetService) startExistingPetsAI() {
	ps.mutex.RLock()
	petsToStart := make([]*models.Pet, 0)
	for _, pet := range ps.pets {
		if pet.IsAlive() {
			petsToStart = append(petsToStart, pet)
		}
	}
	ps.mutex.RUnlock()

	// 在锁外启动AI，避免死锁
	for _, pet := range petsToStart {
		ps.startPetAI(pet)
	}
}

func (ps *PetService) savePetToDatabase(pet *models.Pet) {
	// 更新缓存（确保缓存一致性）
	ps.cacheManager.SetPet(pet.ID, pet)
	ps.cacheManager.SetPetByOwner(pet.Owner, pet)
	
	// 根据存储策略决定如何处理数据
	// 宠物核心数据（等级、金币等）为关键数据
	priority := ps.strategyManager.Strategy.ClassifyPetData(pet, "core")
	
	if ps.strategyManager.Strategy.ShouldPersist(priority) {
		// 使用分层存储策略
		ps.strategyManager.AddData(pet, priority)
	}
	
	// 同时保持原有的批量写入机制
	ps.petRepo.UpdatePetBatch(pet)
}

func (ps *PetService) CreatePet(ownerName string) (*models.Pet, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	for _, existingPet := range ps.pets {
		if existingPet.Owner == ownerName {
			return nil, fmt.Errorf("用户 %s 已经拥有宠物 %s，每位训练师只能拥有一只宠物", ownerName, existingPet.Name)
		}
	}

	if existingPet, err := ps.petRepo.GetPetByOwner(ownerName); err != nil {
		log.Printf("Error checking existing pet in database: %v", err)
	} else if existingPet != nil {
		return nil, fmt.Errorf("用户 %s 已经拥有宠物 %s，每位训练师只能拥有一只宠物", ownerName, existingPet.Name)
	}

	pet := models.NewPet(ownerName)
	
	if err := ps.petRepo.CreatePet(pet); err != nil {
		return nil, fmt.Errorf("failed to save pet to database: %w", err)
	}
	
	ps.pets[pet.ID] = pet

	event := models.Event{
		ID:        uuid.New().String(),
		PetID:     pet.ID,
		PetName:   pet.Name,
		Type:      models.EventExplore,
		Message:   fmt.Sprintf("[%s] 诞生了！准备开始探索世界...", pet.Name),
		Timestamp: time.Now(),
		Data:      models.EventData{Location: pet.Location},
	}

	ps.addEvent(event)
	ps.startPetAI(pet)
	
	return pet, nil
}

func (ps *PetService) GetPet(petID string) (*models.Pet, bool) {
	// 首先尝试从缓存获取
	if pet, exists := ps.cacheManager.GetPet(petID); exists {
		return pet, true
	}
	
	// 缓存未命中，从内存map获取
	ps.mutex.RLock()
	pet, exists := ps.pets[petID]
	ps.mutex.RUnlock()
	
	if exists {
		// 更新缓存
		ps.cacheManager.SetPet(petID, pet)
		return pet, true
	}
	
	// 内存未命中，从数据库获取
	dbPet, err := ps.petRepo.GetPetByID(petID)
	if err != nil {
		log.Printf("Error getting pet from database: %v", err)
		return nil, false
	}
	
	if dbPet != nil {
		// 更新内存和缓存
		ps.mutex.Lock()
		ps.pets[petID] = dbPet
		ps.mutex.Unlock()
		
		ps.cacheManager.SetPet(petID, dbPet)
		ps.cacheManager.SetPetByOwner(dbPet.Owner, dbPet)
		return dbPet, true
	}
	
	return nil, false
}

func (ps *PetService) GetAllPets() []*models.Pet {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	
	pets := make([]*models.Pet, 0, len(ps.pets))
	for _, pet := range ps.pets {
		pets = append(pets, pet)
	}
	return pets
}

func (ps *PetService) GetPetByOwner(ownerName string) (*models.Pet, error) {
	// 首先尝试从缓存获取
	if pet, exists := ps.cacheManager.GetPetByOwner(ownerName); exists {
		return pet, nil
	}
	
	// 缓存未命中，从内存map获取
	ps.mutex.RLock()
	var foundPet *models.Pet
	for _, pet := range ps.pets {
		if pet.Owner == ownerName {
			foundPet = pet
			break
		}
	}
	ps.mutex.RUnlock()
	
	if foundPet != nil {
		// 更新缓存
		ps.cacheManager.SetPetByOwner(ownerName, foundPet)
		return foundPet, nil
	}
	
	// 内存未命中，从数据库获取
	dbPet, err := ps.petRepo.GetPetByOwner(ownerName)
	if err != nil {
		return nil, err
	}
	
	if dbPet != nil {
		// 更新内存和缓存
		ps.mutex.Lock()
		ps.pets[dbPet.ID] = dbPet
		ps.mutex.Unlock()
		
		ps.cacheManager.SetPet(dbPet.ID, dbPet)
		ps.cacheManager.SetPetByOwner(ownerName, dbPet)
	}
	
	return dbPet, nil
}

// GetSystemStats 获取系统统计信息
func (ps *PetService) GetSystemStats() map[string]interface{} {
	cacheStats := ps.cacheManager.GetCacheStats()
	stateStats := ps.stateManager.GetStateStats()
	storageStats := ps.strategyManager.GetStats()
	lockStats := ps.mutex.GetMetrics()
	globalLockStats := utils.GlobalLockManager.GetLockStats()
	
	ps.mutex.RLock()
	petCount := len(ps.pets)
	eventCount := len(ps.events)
	activePetCount := len(ps.activePets)
	ps.mutex.RUnlock()
	
	return map[string]interface{}{
		"pets": map[string]interface{}{
			"total_loaded": petCount,
			"active_ai":    activePetCount,
		},
		"events": map[string]interface{}{
			"in_memory": eventCount,
		},
		"cache":     cacheStats,
		"states":    stateStats,
		"storage":   storageStats,
		"locks": map[string]interface{}{
			"main_mutex": lockStats,
			"global":     globalLockStats,
		},
		"performance": map[string]interface{}{
			"object_pool_enabled": true,
			"json_optimizer_enabled": true,
			"lock_optimization_enabled": true,
		},
	}
}

// SetPetOnlineStatus 设置宠物在线状态
func (ps *PetService) SetPetOnlineStatus(petID string, isOnline bool) {
	ps.stateManager.SetOnlineStatus(petID, isOnline)
}

// GetOnlinePlayerCount 获取在线玩家数量
func (ps *PetService) GetOnlinePlayerCount() int {
	return ps.stateManager.GetOnlinePlayers()
}

// GetPlayerState 获取玩家状态
func (ps *PetService) GetPlayerState(petID string) *cache.PlayerState {
	return ps.stateManager.GetPlayerState(petID)
}

// LogPerformanceStats 定期记录性能统计
func (ps *PetService) LogPerformanceStats() {
	ps.cacheManager.LogCacheStats()
	log.Printf("System Performance Stats: %+v", ps.GetSystemStats())
}