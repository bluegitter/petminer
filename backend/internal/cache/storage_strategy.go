package cache

import (
	"log"
	"miningpet/internal/models"
	"sync"
	"time"
)

// DataPriority 数据优先级
type DataPriority int

const (
	// Critical 关键数据 - 必须持久化，不能丢失
	Critical DataPriority = iota
	// Important 重要数据 - 应该持久化，短期丢失可接受
	Important
	// Transient 瞬时数据 - 可以丢失，仅内存存储
	Transient
)

// StorageLayer 存储层
type StorageLayer int

const (
	MemoryOnly StorageLayer = iota  // 仅内存
	MemoryAndCache                  // 内存+缓存
	MemoryAndDB                     // 内存+数据库
	AllLayers                       // 内存+缓存+数据库
)

// DataClassification 数据分类规则
type DataClassification struct {
	// 宠物相关数据
	PetCoreData      DataPriority // 宠物基础数据（等级、金币等）
	PetStatusData    DataPriority // 宠物状态数据（血量、位置等）
	PetTempData      DataPriority // 宠物临时数据（buff等）
	
	// 事件相关数据
	CriticalEvents   DataPriority // 关键事件（死亡、升级等）
	NormalEvents     DataPriority // 普通事件（探索、战斗等）
	LogEvents        DataPriority // 日志事件（聊天、系统消息等）
	
	// 系统相关数据
	PlayerSession    DataPriority // 玩家会话数据
	SystemStats      DataPriority // 系统统计数据
	TempBuffs        DataPriority // 临时buff数据
}

// DefaultClassification 默认数据分类
var DefaultClassification = DataClassification{
	// 宠物数据
	PetCoreData:   Critical,   // 等级、金币、物品等核心数据
	PetStatusData: Important,  // 血量、位置等状态数据
	PetTempData:   Transient,  // 临时buff等
	
	// 事件数据
	CriticalEvents: Critical,   // 死亡、升级等关键事件
	NormalEvents:   Important,  // 普通游戏事件
	LogEvents:      Transient,  // 日志类事件
	
	// 系统数据
	PlayerSession: Transient,   // 在线状态等会话数据
	SystemStats:   Important,   // 系统统计数据
	TempBuffs:     Transient,   // 临时buff
}

// StorageStrategy 存储策略
type StorageStrategy struct {
	classification DataClassification
	
	// 不同优先级数据的存储层策略
	priorityMapping map[DataPriority]StorageLayer
	
	// 数据生命周期管理
	lifeCycleRules map[DataPriority]time.Duration
	
	// 批量写入阈值
	batchThresholds map[DataPriority]int
	
	// 刷新间隔
	flushIntervals map[DataPriority]time.Duration
}

// NewStorageStrategy 创建存储策略
func NewStorageStrategy() *StorageStrategy {
	return &StorageStrategy{
		classification: DefaultClassification,
		
		priorityMapping: map[DataPriority]StorageLayer{
			Critical:  AllLayers,      // 关键数据存储在所有层
			Important: MemoryAndDB,    // 重要数据存储在内存和数据库
			Transient: MemoryOnly,     // 瞬时数据仅存储在内存
		},
		
		lifeCycleRules: map[DataPriority]time.Duration{
			Critical:  0,                // 关键数据永不过期
			Important: 24 * time.Hour,   // 重要数据24小时过期
			Transient: 30 * time.Minute, // 瞬时数据30分钟过期
		},
		
		batchThresholds: map[DataPriority]int{
			Critical:  1,   // 关键数据立即写入
			Important: 20,  // 重要数据20条批量写入
			Transient: 100, // 瞬时数据100条批量写入（虽然不写DB）
		},
		
		flushIntervals: map[DataPriority]time.Duration{
			Critical:  time.Second,      // 关键数据1秒刷新
			Important: 30 * time.Second, // 重要数据30秒刷新
			Transient: 5 * time.Minute,  // 瞬时数据5分钟刷新
		},
	}
}

// ClassifyPetData 分类宠物数据
func (ss *StorageStrategy) ClassifyPetData(pet *models.Pet, changeType string) DataPriority {
	switch changeType {
	case "level", "experience", "coins", "items":
		return ss.classification.PetCoreData
	case "health", "energy", "hunger", "location", "status":
		return ss.classification.PetStatusData
	case "temp_buff", "mood":
		return ss.classification.PetTempData
	default:
		return ss.classification.PetStatusData
	}
}

// ClassifyEventData 分类事件数据
func (ss *StorageStrategy) ClassifyEventData(event *models.Event) DataPriority {
	switch event.Type {
	case models.EventRareFind: // 使用现有的稀有发现事件作为关键事件
		return ss.classification.CriticalEvents
	case models.EventBattle, models.EventDiscovery, models.EventExplore:
		return ss.classification.NormalEvents
	case models.EventSocial:
		return ss.classification.LogEvents
	default:
		return ss.classification.NormalEvents
	}
}

// GetStorageLayer 获取数据应该存储的层
func (ss *StorageStrategy) GetStorageLayer(priority DataPriority) StorageLayer {
	if layer, exists := ss.priorityMapping[priority]; exists {
		return layer
	}
	return MemoryOnly
}

// GetLifeCycle 获取数据生命周期
func (ss *StorageStrategy) GetLifeCycle(priority DataPriority) time.Duration {
	if duration, exists := ss.lifeCycleRules[priority]; exists {
		return duration
	}
	return 30 * time.Minute // 默认30分钟
}

// GetBatchThreshold 获取批量写入阈值
func (ss *StorageStrategy) GetBatchThreshold(priority DataPriority) int {
	if threshold, exists := ss.batchThresholds[priority]; exists {
		return threshold
	}
	return 50 // 默认50条
}

// GetFlushInterval 获取刷新间隔
func (ss *StorageStrategy) GetFlushInterval(priority DataPriority) time.Duration {
	if interval, exists := ss.flushIntervals[priority]; exists {
		return interval
	}
	return 1 * time.Minute // 默认1分钟
}

// ShouldPersist 判断是否应该持久化
func (ss *StorageStrategy) ShouldPersist(priority DataPriority) bool {
	layer := ss.GetStorageLayer(priority)
	return layer == MemoryAndDB || layer == AllLayers
}

// ShouldCache 判断是否应该缓存
func (ss *StorageStrategy) ShouldCache(priority DataPriority) bool {
	layer := ss.GetStorageLayer(priority)
	return layer == MemoryAndCache || layer == AllLayers
}

// StrategyManager 策略管理器
type StrategyManager struct {
	Strategy *StorageStrategy
	
	// 分层数据队列
	criticalQueue  chan interface{}
	importantQueue chan interface{}
	transientQueue chan interface{}
	
	// 队列处理器
	processors map[DataPriority]*QueueProcessor
	
	// 统计信息
	stats map[DataPriority]*StorageStats
	mutex sync.RWMutex
}

// StorageStats 存储统计
type StorageStats struct {
	ProcessedCount int64
	ErrorCount     int64
	LastProcessed  time.Time
	QueueLength    int
}

// QueueProcessor 队列处理器
type QueueProcessor struct {
	priority    DataPriority
	queue      chan interface{}
	batchSize  int
	flushTime  time.Duration
	stopChan   chan bool
	processFunc func([]interface{}) error
}

// NewStrategyManager 创建策略管理器
func NewStrategyManager() *StrategyManager {
	strategy := NewStorageStrategy()
	
	sm := &StrategyManager{
		Strategy: strategy,
		
		criticalQueue:  make(chan interface{}, 100),
		importantQueue: make(chan interface{}, 500),
		transientQueue: make(chan interface{}, 1000),
		
		processors: make(map[DataPriority]*QueueProcessor),
		stats:      make(map[DataPriority]*StorageStats),
	}
	
	// 初始化处理器
	sm.initProcessors()
	
	return sm
}

// initProcessors 初始化处理器
func (sm *StrategyManager) initProcessors() {
	// 关键数据处理器
	sm.processors[Critical] = &QueueProcessor{
		priority:  Critical,
		queue:     sm.criticalQueue,
		batchSize: sm.Strategy.GetBatchThreshold(Critical),
		flushTime: sm.Strategy.GetFlushInterval(Critical),
		stopChan:  make(chan bool),
		processFunc: sm.processCriticalData,
	}
	
	// 重要数据处理器
	sm.processors[Important] = &QueueProcessor{
		priority:  Important,
		queue:     sm.importantQueue,
		batchSize: sm.Strategy.GetBatchThreshold(Important),
		flushTime: sm.Strategy.GetFlushInterval(Important),
		stopChan:  make(chan bool),
		processFunc: sm.processImportantData,
	}
	
	// 瞬时数据处理器
	sm.processors[Transient] = &QueueProcessor{
		priority:  Transient,
		queue:     sm.transientQueue,
		batchSize: sm.Strategy.GetBatchThreshold(Transient),
		flushTime: sm.Strategy.GetFlushInterval(Transient),
		stopChan:  make(chan bool),
		processFunc: sm.processTransientData,
	}
	
	// 初始化统计
	for priority := range sm.processors {
		sm.stats[priority] = &StorageStats{}
	}
	
	// 启动处理器
	for _, processor := range sm.processors {
		go processor.run()
	}
}

// AddData 添加数据到相应队列
func (sm *StrategyManager) AddData(data interface{}, priority DataPriority) {
	var queue chan interface{}
	
	switch priority {
	case Critical:
		queue = sm.criticalQueue
	case Important:
		queue = sm.importantQueue
	case Transient:
		queue = sm.transientQueue
	default:
		queue = sm.importantQueue
	}
	
	select {
	case queue <- data:
		// 更新队列长度统计
		sm.mutex.Lock()
		sm.stats[priority].QueueLength = len(queue)
		sm.mutex.Unlock()
	default:
		log.Printf("Warning: %v priority queue is full, dropping data", priority)
		sm.mutex.Lock()
		sm.stats[priority].ErrorCount++
		sm.mutex.Unlock()
	}
}

// processCriticalData 处理关键数据
func (sm *StrategyManager) processCriticalData(items []interface{}) error {
	log.Printf("Processing %d critical data items", len(items))
	
	// 关键数据立即持久化到数据库
	for _, item := range items {
		// 实际的数据库写入逻辑
		_ = item
	}
	
	sm.mutex.Lock()
	sm.stats[Critical].ProcessedCount += int64(len(items))
	sm.stats[Critical].LastProcessed = time.Now()
	sm.mutex.Unlock()
	
	return nil
}

// processImportantData 处理重要数据
func (sm *StrategyManager) processImportantData(items []interface{}) error {
	log.Printf("Processing %d important data items", len(items))
	
	// 重要数据批量持久化
	for _, item := range items {
		// 实际的批量写入逻辑
		_ = item
	}
	
	sm.mutex.Lock()
	sm.stats[Important].ProcessedCount += int64(len(items))
	sm.stats[Important].LastProcessed = time.Now()
	sm.mutex.Unlock()
	
	return nil
}

// processTransientData 处理瞬时数据
func (sm *StrategyManager) processTransientData(items []interface{}) error {
	log.Printf("Processing %d transient data items (memory only)", len(items))
	
	// 瞬时数据仅在内存中处理，不持久化
	sm.mutex.Lock()
	sm.stats[Transient].ProcessedCount += int64(len(items))
	sm.stats[Transient].LastProcessed = time.Now()
	sm.mutex.Unlock()
	
	return nil
}

// run 运行队列处理器
func (qp *QueueProcessor) run() {
	ticker := time.NewTicker(qp.flushTime)
	defer ticker.Stop()
	
	var batch []interface{}
	
	flush := func() {
		if len(batch) == 0 {
			return
		}
		
		if err := qp.processFunc(batch); err != nil {
			log.Printf("Error processing %v priority data: %v", qp.priority, err)
		}
		
		batch = batch[:0] // 清空但保留容量
	}
	
	for {
		select {
		case item := <-qp.queue:
			batch = append(batch, item)
			if len(batch) >= qp.batchSize {
				flush()
			}
			
		case <-ticker.C:
			flush()
			
		case <-qp.stopChan:
			flush() // 最后一次刷新
			return
		}
	}
}

// GetStats 获取存储统计
func (sm *StrategyManager) GetStats() map[DataPriority]*StorageStats {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	// 更新队列长度
	sm.stats[Critical].QueueLength = len(sm.criticalQueue)
	sm.stats[Important].QueueLength = len(sm.importantQueue)
	sm.stats[Transient].QueueLength = len(sm.transientQueue)
	
	// 返回副本
	stats := make(map[DataPriority]*StorageStats)
	for priority, stat := range sm.stats {
		statCopy := *stat
		stats[priority] = &statCopy
	}
	
	return stats
}

// Stop 停止策略管理器
func (sm *StrategyManager) Stop() {
	for _, processor := range sm.processors {
		close(processor.stopChan)
	}
	log.Println("Storage strategy manager stopped")
}