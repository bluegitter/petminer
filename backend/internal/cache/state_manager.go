package cache

import (
	"log"
	"sync"
	"time"
)

// PlayerState 玩家状态（轻量级，高频读写）
type PlayerState struct {
	PetID        string    `json:"pet_id"`
	LastActivity time.Time `json:"last_activity"`
	IsOnline     bool      `json:"is_online"`
	CurrentHP    int       `json:"current_hp"`
	CurrentMP    int       `json:"current_mp"`
	Location     string    `json:"location"`
	LastAction   string    `json:"last_action"`
	// 临时状态，不需要持久化
	TempBuff     map[string]interface{} `json:"temp_buff,omitempty"`
	ActionCount  int                    `json:"action_count"`
}

// StateManager 状态管理器 - 内存优先策略
type StateManager struct {
	states       map[string]*PlayerState // petID -> PlayerState
	mutex        sync.RWMutex
	cache        *StatsMemoryCache
	
	// 状态变更队列（用于批量持久化关键状态）
	changeQueue  chan StateChange
	flushTicker  *time.Ticker
	stopChan     chan bool
}

// StateChange 状态变更记录
type StateChange struct {
	PetID     string
	Field     string
	OldValue  interface{}
	NewValue  interface{}
	Timestamp time.Time
	IsCritical bool // 是否为关键状态（需要立即持久化）
}

// NewStateManager 创建状态管理器
func NewStateManager() *StateManager {
	sm := &StateManager{
		states:      make(map[string]*PlayerState),
		cache:       NewStatsMemoryCache(10*time.Minute, 2*time.Minute), // 状态缓存
		changeQueue: make(chan StateChange, 1000),
		flushTicker: time.NewTicker(30 * time.Second), // 每30秒批量持久化
		stopChan:    make(chan bool),
	}
	
	go sm.processStateChanges()
	return sm
}

// GetPlayerState 获取玩家状态（内存优先）
func (sm *StateManager) GetPlayerState(petID string) *PlayerState {
	sm.mutex.RLock()
	state, exists := sm.states[petID]
	sm.mutex.RUnlock()
	
	if exists {
		return state
	}
	
	// 内存中不存在，检查缓存
	if cached, exists := sm.cache.Get("state:" + petID); exists {
		if state, ok := cached.(*PlayerState); ok {
			sm.mutex.Lock()
			sm.states[petID] = state
			sm.mutex.Unlock()
			return state
		}
	}
	
	// 创建新状态
	state = &PlayerState{
		PetID:        petID,
		LastActivity: time.Now(),
		IsOnline:     false,
		TempBuff:     make(map[string]interface{}),
		ActionCount:  0,
	}
	
	sm.mutex.Lock()
	sm.states[petID] = state
	sm.mutex.Unlock()
	
	sm.cache.Set("state:"+petID, state, 0)
	return state
}

// UpdatePlayerState 更新玩家状态
func (sm *StateManager) UpdatePlayerState(petID string, updateFunc func(*PlayerState) []StateChange) {
	sm.mutex.Lock()
	state, exists := sm.states[petID]
	if !exists {
		state = sm.GetPlayerState(petID)
	}
	
	// 执行更新并获取变更记录
	changes := updateFunc(state)
	state.LastActivity = time.Now()
	
	sm.mutex.Unlock()
	
	// 记录变更
	for _, change := range changes {
		change.Timestamp = time.Now()
		select {
		case sm.changeQueue <- change:
		default:
			log.Printf("Warning: state change queue is full, dropping change for pet %s", petID)
		}
	}
	
	// 更新缓存
	sm.cache.Set("state:"+petID, state, 0)
}

// SetOnlineStatus 设置在线状态
func (sm *StateManager) SetOnlineStatus(petID string, isOnline bool) {
	sm.UpdatePlayerState(petID, func(state *PlayerState) []StateChange {
		oldStatus := state.IsOnline
		state.IsOnline = isOnline
		
		return []StateChange{{
			PetID:      petID,
			Field:      "is_online",
			OldValue:   oldStatus,
			NewValue:   isOnline,
			IsCritical: false, // 在线状态不是关键状态
		}}
	})
}

// UpdateHP 更新血量
func (sm *StateManager) UpdateHP(petID string, newHP int) {
	sm.UpdatePlayerState(petID, func(state *PlayerState) []StateChange {
		oldHP := state.CurrentHP
		state.CurrentHP = newHP
		
		return []StateChange{{
			PetID:      petID,
			Field:      "current_hp",
			OldValue:   oldHP,
			NewValue:   newHP,
			IsCritical: newHP <= 0, // 血量为0是关键状态
		}}
	})
}

// UpdateLocation 更新位置
func (sm *StateManager) UpdateLocation(petID string, location string) {
	sm.UpdatePlayerState(petID, func(state *PlayerState) []StateChange {
		oldLocation := state.Location
		state.Location = location
		
		return []StateChange{{
			PetID:      petID,
			Field:      "location",
			OldValue:   oldLocation,
			NewValue:   location,
			IsCritical: false,
		}}
	})
}

// IncrementActionCount 增加行动计数
func (sm *StateManager) IncrementActionCount(petID string) {
	sm.UpdatePlayerState(petID, func(state *PlayerState) []StateChange {
		oldCount := state.ActionCount
		state.ActionCount++
		
		return []StateChange{{
			PetID:      petID,
			Field:      "action_count",
			OldValue:   oldCount,
			NewValue:   state.ActionCount,
			IsCritical: false,
		}}
	})
}

// SetTempBuff 设置临时buff
func (sm *StateManager) SetTempBuff(petID string, buffName string, buffValue interface{}) {
	sm.UpdatePlayerState(petID, func(state *PlayerState) []StateChange {
		if state.TempBuff == nil {
			state.TempBuff = make(map[string]interface{})
		}
		oldValue := state.TempBuff[buffName]
		state.TempBuff[buffName] = buffValue
		
		return []StateChange{{
			PetID:      petID,
			Field:      "temp_buff." + buffName,
			OldValue:   oldValue,
			NewValue:   buffValue,
			IsCritical: false, // 临时buff不是关键状态
		}}
	})
}

// RemoveTempBuff 移除临时buff
func (sm *StateManager) RemoveTempBuff(petID string, buffName string) {
	sm.UpdatePlayerState(petID, func(state *PlayerState) []StateChange {
		if state.TempBuff == nil {
			return []StateChange{}
		}
		
		oldValue := state.TempBuff[buffName]
		delete(state.TempBuff, buffName)
		
		return []StateChange{{
			PetID:      petID,
			Field:      "temp_buff." + buffName,
			OldValue:   oldValue,
			NewValue:   nil,
			IsCritical: false,
		}}
	})
}

// GetOnlinePlayers 获取在线玩家数量
func (sm *StateManager) GetOnlinePlayers() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	count := 0
	for _, state := range sm.states {
		if state.IsOnline {
			count++
		}
	}
	return count
}

// GetActiveStates 获取活跃状态（最近活动的玩家）
func (sm *StateManager) GetActiveStates(since time.Duration) map[string]*PlayerState {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	threshold := time.Now().Add(-since)
	active := make(map[string]*PlayerState)
	
	for petID, state := range sm.states {
		if state.LastActivity.After(threshold) {
			active[petID] = state
		}
	}
	
	return active
}

// CleanupInactiveStates 清理非活跃状态
func (sm *StateManager) CleanupInactiveStates(inactiveThreshold time.Duration) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	threshold := time.Now().Add(-inactiveThreshold)
	cleaned := 0
	
	for petID, state := range sm.states {
		if state.LastActivity.Before(threshold) && !state.IsOnline {
			delete(sm.states, petID)
			sm.cache.Delete("state:" + petID)
			cleaned++
		}
	}
	
	if cleaned > 0 {
		log.Printf("Cleaned up %d inactive player states", cleaned)
	}
}

// GetStateStats 获取状态管理统计
func (sm *StateManager) GetStateStats() map[string]interface{} {
	sm.mutex.RLock()
	totalStates := len(sm.states)
	sm.mutex.RUnlock()
	
	onlineCount := sm.GetOnlinePlayers()
	activeCount := len(sm.GetActiveStates(5 * time.Minute))
	cacheStats := sm.cache.GetStats()
	
	return map[string]interface{}{
		"total_states":  totalStates,
		"online_count":  onlineCount,
		"active_count":  activeCount,
		"cache_stats":   cacheStats,
		"queue_length":  len(sm.changeQueue),
	}
}

// processStateChanges 处理状态变更（后台协程）
func (sm *StateManager) processStateChanges() {
	criticalChanges := make([]StateChange, 0)
	regularChanges := make([]StateChange, 0)
	
	processBatch := func() {
		if len(criticalChanges) > 0 {
			// 立即处理关键变更
			log.Printf("Processing %d critical state changes", len(criticalChanges))
			criticalChanges = criticalChanges[:0]
		}
		
		if len(regularChanges) > 0 {
			// 批量处理常规变更
			log.Printf("Processing %d regular state changes", len(regularChanges))
			regularChanges = regularChanges[:0]
		}
	}
	
	for {
		select {
		case change := <-sm.changeQueue:
			if change.IsCritical {
				criticalChanges = append(criticalChanges, change)
				if len(criticalChanges) >= 10 { // 关键变更积累10个就处理
					processBatch()
				}
			} else {
				regularChanges = append(regularChanges, change)
				if len(regularChanges) >= 100 { // 常规变更积累100个再处理
					processBatch()
				}
			}
			
		case <-sm.flushTicker.C:
			processBatch()
			// 定期清理非活跃状态
			sm.CleanupInactiveStates(30 * time.Minute)
			
		case <-sm.stopChan:
			processBatch() // 最后一次处理
			return
		}
	}
}

// Stop 停止状态管理器
func (sm *StateManager) Stop() {
	close(sm.stopChan)
	sm.flushTicker.Stop()
	sm.cache.Stop()
	log.Println("State manager stopped")
}