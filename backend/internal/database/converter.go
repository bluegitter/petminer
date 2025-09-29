package database

import (
	"miningpet/internal/models"
	"time"
)

// ConvertToDBPet 将内存宠物模型转换为数据库模型
func ConvertToDBPet(pet *models.Pet) (*DBPet, error) {
	dbPet := &DBPet{
		ID:           pet.ID,
		Name:         pet.Name,
		Owner:        pet.Owner,
		Personality:  string(pet.Personality),
		Level:        pet.Level,
		Experience:   pet.Experience,
		Health:       pet.Health,
		MaxHealth:    pet.MaxHealth,
		Energy:       pet.Energy,
		MaxEnergy:    pet.MaxEnergy,
		Hunger:       pet.Hunger,
		Social:       pet.Social,
		Mood:         string(pet.Mood),
		Attack:       pet.Attack,
		Defense:      pet.Defense,
		Coins:        pet.Coins,
		Location:     pet.Location,
		Status:       string(pet.Status),
		LastActivity: pet.LastActivity,
		CreatedAt:    pet.CreatedAt,
		UpdatedAt:    time.Now(),
	}

	// 设置JSON字段
	if err := dbPet.SetMemory(pet.Memory); err != nil {
		return nil, err
	}

	if err := dbPet.SetFriends(pet.Friends); err != nil {
		return nil, err
	}

	return dbPet, nil
}

// ConvertFromDBPet 将数据库模型转换为内存宠物模型
func ConvertFromDBPet(dbPet *DBPet) (*models.Pet, error) {
	memory, err := dbPet.GetMemory()
	if err != nil {
		return nil, err
	}

	friends, err := dbPet.GetFriends()
	if err != nil {
		return nil, err
	}

	pet := &models.Pet{
		ID:           dbPet.ID,
		Name:         dbPet.Name,
		Owner:        dbPet.Owner,
		Personality:  models.PetPersonality(dbPet.Personality),
		Level:        dbPet.Level,
		Experience:   dbPet.Experience,
		Health:       dbPet.Health,
		MaxHealth:    dbPet.MaxHealth,
		Energy:       dbPet.Energy,
		MaxEnergy:    dbPet.MaxEnergy,
		Hunger:       dbPet.Hunger,
		Social:       dbPet.Social,
		Mood:         models.PetMood(dbPet.Mood),
		Attack:       dbPet.Attack,
		Defense:      dbPet.Defense,
		Coins:        dbPet.Coins,
		Location:     dbPet.Location,
		Status:       models.PetStatus(dbPet.Status),
		Memory:       memory,
		Friends:      friends,
		LastActivity: dbPet.LastActivity,
		CreatedAt:    dbPet.CreatedAt,
	}

	return pet, nil
}

// ConvertToDBEvent 将内存事件模型转换为数据库模型
func ConvertToDBEvent(event *models.Event) (*DBEvent, error) {
	dbEvent := &DBEvent{
		ID:        event.ID,
		PetID:     event.PetID,
		PetName:   event.PetName,
		Type:      string(event.Type),
		Message:   event.Message,
		Timestamp: event.Timestamp,
		CreatedAt: time.Now(),
	}

	// 设置事件数据
	if err := dbEvent.SetEventData(event.Data); err != nil {
		return nil, err
	}

	return dbEvent, nil
}

// ConvertFromDBEvent 将数据库模型转换为内存事件模型
func ConvertFromDBEvent(dbEvent *DBEvent) (*models.Event, error) {
	eventDataMap, err := dbEvent.GetEventData()
	if err != nil {
		return nil, err
	}

	// 将map转换为EventData结构
	eventData := models.EventData{}
	
	if location, ok := eventDataMap["location"].(string); ok {
		eventData.Location = location
	}
	if experience, ok := eventDataMap["experience"].(float64); ok {
		eventData.Experience = int(experience)
	}
	if coins, ok := eventDataMap["coins"].(float64); ok {
		eventData.Coins = int(coins)
	}
	if enemy, ok := eventDataMap["enemy"].(string); ok {
		eventData.Enemy = enemy
	}
	if damage, ok := eventDataMap["damage"].(float64); ok {
		eventData.Damage = int(damage)
	}
	if isVictory, ok := eventDataMap["is_victory"].(bool); ok {
		eventData.IsVictory = isVictory
	}
	if friendName, ok := eventDataMap["friend_name"].(string); ok {
		eventData.FriendName = friendName
	}
	if newLevel, ok := eventDataMap["new_level"].(float64); ok {
		eventData.NewLevel = int(newLevel)
	}
	if rareItem, ok := eventDataMap["rare_item"].(string); ok {
		eventData.RareItem = rareItem
	}

	event := &models.Event{
		ID:        dbEvent.ID,
		PetID:     dbEvent.PetID,
		PetName:   dbEvent.PetName,
		Type:      models.EventType(dbEvent.Type),
		Message:   dbEvent.Message,
		Timestamp: dbEvent.Timestamp,
		Data:      eventData,
	}

	return event, nil
}