package database

import (
	"fmt"
	"log"
	"miningpet/internal/models"
	"time"

	"gorm.io/gorm"
)

// PetRepository 宠物数据访问层
type PetRepository struct {
	db *gorm.DB
}

// NewPetRepository 创建宠物仓库
func NewPetRepository() *PetRepository {
	return &PetRepository{db: DB}
}

// CreatePet 创建宠物
func (r *PetRepository) CreatePet(pet *models.Pet) error {
	dbPet, err := ConvertToDBPet(pet)
	if err != nil {
		return fmt.Errorf("failed to convert pet: %w", err)
	}

	if err := r.db.Create(dbPet).Error; err != nil {
		return fmt.Errorf("failed to create pet: %w", err)
	}

	return nil
}

// GetPetByID 根据ID获取宠物
func (r *PetRepository) GetPetByID(id string) (*models.Pet, error) {
	var dbPet DBPet
	if err := r.db.Where("id = ?", id).First(&dbPet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pet: %w", err)
	}

	pet, err := ConvertFromDBPet(&dbPet)
	if err != nil {
		return nil, fmt.Errorf("failed to convert pet: %w", err)
	}

	return pet, nil
}

// GetPetByOwner 根据主人获取宠物
func (r *PetRepository) GetPetByOwner(owner string) (*models.Pet, error) {
	var dbPet DBPet
	if err := r.db.Where("owner = ?", owner).First(&dbPet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pet by owner: %w", err)
	}

	pet, err := ConvertFromDBPet(&dbPet)
	if err != nil {
		return nil, fmt.Errorf("failed to convert pet: %w", err)
	}

	return pet, nil
}

// GetAllPets 获取所有宠物
func (r *PetRepository) GetAllPets() ([]*models.Pet, error) {
	var dbPets []DBPet
	if err := r.db.Find(&dbPets).Error; err != nil {
		return nil, fmt.Errorf("failed to get all pets: %w", err)
	}

	pets := make([]*models.Pet, len(dbPets))
	for i, dbPet := range dbPets {
		pet, err := ConvertFromDBPet(&dbPet)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pet: %w", err)
		}
		pets[i] = pet
	}

	return pets, nil
}

// UpdatePet 更新宠物
func (r *PetRepository) UpdatePet(pet *models.Pet) error {
	dbPet, err := ConvertToDBPet(pet)
	if err != nil {
		return fmt.Errorf("failed to convert pet: %w", err)
	}

	if err := r.db.Where("id = ?", pet.ID).Updates(dbPet).Error; err != nil {
		return fmt.Errorf("failed to update pet: %w", err)
	}

	return nil
}

// UpdatePetBatch 批量更新宠物（高性能）
func (r *PetRepository) UpdatePetBatch(pet *models.Pet) {
	if PetBatchManager != nil {
		PetBatchManager.AddWrite(&PetBatchWrite{Pet: pet})
	} else {
		// 降级到同步写入
		if err := r.UpdatePet(pet); err != nil {
			log.Printf("Failed to update pet: %v", err)
		}
	}
}

// DeletePet 删除宠物
func (r *PetRepository) DeletePet(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&DBPet{}).Error; err != nil {
		return fmt.Errorf("failed to delete pet: %w", err)
	}

	return nil
}

// EventRepository 事件数据访问层
type EventRepository struct {
	db *gorm.DB
}

// NewEventRepository 创建事件仓库
func NewEventRepository() *EventRepository {
	return &EventRepository{db: DB}
}

// CreateEvent 创建事件
func (r *EventRepository) CreateEvent(event *models.Event) error {
	dbEvent, err := ConvertToDBEvent(event)
	if err != nil {
		return fmt.Errorf("failed to convert event: %w", err)
	}

	if err := r.db.Create(dbEvent).Error; err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// CreateEventBatch 批量创建事件（高性能）
func (r *EventRepository) CreateEventBatch(event *models.Event) {
	if EventBatchManager != nil {
		EventBatchManager.AddWrite(&EventBatchWrite{Event: event})
	} else {
		// 降级到同步写入
		if err := r.CreateEvent(event); err != nil {
			log.Printf("Failed to create event: %v", err)
		}
	}
}

// GetEventsByPetID 根据宠物ID获取事件
func (r *EventRepository) GetEventsByPetID(petID string, limit int) ([]*models.Event, error) {
	var dbEvents []DBEvent
	query := r.db.Where("pet_id = ?", petID).Order("timestamp DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&dbEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	events := make([]*models.Event, len(dbEvents))
	for i, dbEvent := range dbEvents {
		event, err := ConvertFromDBEvent(&dbEvent)
		if err != nil {
			return nil, fmt.Errorf("failed to convert event: %w", err)
		}
		events[i] = event
	}

	return events, nil
}

// GetRecentEvents 获取最近的事件
func (r *EventRepository) GetRecentEvents(limit int) ([]*models.Event, error) {
	var dbEvents []DBEvent
	query := r.db.Order("timestamp DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&dbEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent events: %w", err)
	}

	events := make([]*models.Event, len(dbEvents))
	for i, dbEvent := range dbEvents {
		event, err := ConvertFromDBEvent(&dbEvent)
		if err != nil {
			return nil, fmt.Errorf("failed to convert event: %w", err)
		}
		events[i] = event
	}

	return events, nil
}

// DeleteOldEvents 删除指定时间之前的事件
func (r *EventRepository) DeleteOldEvents(before time.Time) error {
	if err := r.db.Where("timestamp < ?", before).Delete(&DBEvent{}).Error; err != nil {
		return fmt.Errorf("failed to delete old events: %w", err)
	}

	return nil
}

// GetEventCount 获取事件总数
func (r *EventRepository) GetEventCount() (int64, error) {
	var count int64
	if err := r.db.Model(&DBEvent{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}

// EventBatchWrite 事件批量写入操作
type EventBatchWrite struct {
	Event *models.Event
}

// Execute 执行事件批量写入
func (ebw *EventBatchWrite) Execute(tx *gorm.DB) error {
	dbEvent, err := ConvertToDBEvent(ebw.Event)
	if err != nil {
		return fmt.Errorf("failed to convert event: %w", err)
	}

	if err := tx.Create(dbEvent).Error; err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// PetBatchWrite 宠物批量写入操作
type PetBatchWrite struct {
	Pet *models.Pet
}

// Execute 执行宠物批量写入
func (pbw *PetBatchWrite) Execute(tx *gorm.DB) error {
	dbPet, err := ConvertToDBPet(pbw.Pet)
	if err != nil {
		return fmt.Errorf("failed to convert pet: %w", err)
	}

	if err := tx.Where("id = ?", pbw.Pet.ID).Updates(dbPet).Error; err != nil {
		return fmt.Errorf("failed to update pet: %w", err)
	}

	return nil
}

// 全局批量写入管理器
var (
	EventBatchManager *BatchWriteManager
	PetBatchManager   *BatchWriteManager
)

// InitializeBatchManagers 初始化批量写入管理器
func InitializeBatchManagers() {
	// 事件批量写入：每50条或每2秒刷新一次
	EventBatchManager = NewBatchWriteManager(50, 2*time.Second)
	
	// 宠物批量写入：每20条或每5秒刷新一次（宠物更新频率较低）
	PetBatchManager = NewBatchWriteManager(20, 5*time.Second)
}

// CloseBatchManagers 关闭批量写入管理器
func CloseBatchManagers() {
	if EventBatchManager != nil {
		EventBatchManager.Stop()
	}
	if PetBatchManager != nil {
		PetBatchManager.Stop()
	}
}