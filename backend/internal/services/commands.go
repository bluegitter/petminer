package services

import (
	"fmt"
	"time"

	"miningpet/internal/models"
	"github.com/google/uuid"
)

func (ps *PetService) ExecuteCommand(petID, command string, params map[string]interface{}) (interface{}, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	pet, exists := ps.pets[petID]
	if !exists {
		return nil, fmt.Errorf("pet not found")
	}

	switch command {
	case "rest":
		return ps.executeRestCommand(pet, params)
	case "feed":
		return ps.executeFeedCommand(pet, params)
	case "socialize":
		return ps.executeSocializeCommand(pet, params)
	case "explore":
		return ps.executeExploreCommand(pet, params)
	case "addcoins":
		return ps.executeAddCoinsCommand(pet, params)
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

func (ps *PetService) RestPet(petID string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	pet, exists := ps.pets[petID]
	if !exists {
		return fmt.Errorf("pet not found")
	}

	if !pet.CanRest() {
		if !pet.IsAlive() {
			return fmt.Errorf("宠物已倒下，无法休息（生命值: %d）", pet.Health)
		}
		if pet.Status != models.StatusIdle {
			return fmt.Errorf("宠物当前状态为 %s，无法休息", pet.Status)
		}
		return fmt.Errorf("宠物暂时无法休息（状态: %s，生命值: %d）", pet.Status, pet.Health)
	}

	action := Action{
		Type:     ActionRest,
		Priority: 100,
		Reason:   "主人命令休息",
		Duration: 30,
	}

	ps.executeRestAction(pet, action)
	return nil
}

func (ps *PetService) FeedPet(petID string, amount int) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	pet, exists := ps.pets[petID]
	if !exists {
		return fmt.Errorf("pet not found")
	}

	if !pet.IsAlive() {
		return fmt.Errorf("pet is not alive")
	}

	if amount <= 0 {
		amount = 20
	}

	cost := amount / 2
	if pet.Coins < cost {
		return fmt.Errorf("not enough coins to feed pet")
	}

	pet.Coins -= cost
	pet.Feed(amount)

	event := models.Event{
		ID:        uuid.New().String(),
		PetID:     pet.ID,
		PetName:   pet.Name,
		Type:      models.EventReward,
		Message:   fmt.Sprintf("[%s] 主人喂食了%d点，消耗%d金币", pet.Name, amount, cost),
		Timestamp: time.Now(),
		Data:      models.EventData{Coins: -cost},
	}
	ps.addEvent(event)

	return nil
}

func (ps *PetService) SocializePet(petID string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	pet, exists := ps.pets[petID]
	if !exists {
		return fmt.Errorf("pet not found")
	}

	if !pet.CanSocialize() {
		if !pet.IsAlive() {
			return fmt.Errorf("宠物已倒下，无法社交（生命值: %d）", pet.Health)
		}
		if pet.Status != "等待中" {
			return fmt.Errorf("宠物当前状态为 %s，无法社交", pet.Status)
		}
		if pet.Social >= 90 {
			return fmt.Errorf("宠物社交需求已满足（社交度: %d/100），暂时不需要社交", pet.Social)
		}
		return fmt.Errorf("宠物暂时无法社交（状态: %s，社交度: %d，生命值: %d）", pet.Status, pet.Social, pet.Health)
	}

	action := Action{
		Type:     ActionSocialize,
		Priority: 100,
		Reason:   "主人安排社交",
		Duration: 40,
	}

	ps.executeSocializeAction(pet, action)
	return nil
}

func (ps *PetService) GetPetStatus(petID string) (map[string]interface{}, error) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	pet, exists := ps.pets[petID]
	if !exists {
		return nil, fmt.Errorf("pet not found")
	}

	status := map[string]interface{}{
		"basic_info": map[string]interface{}{
			"id":           pet.ID,
			"name":         pet.Name,
			"owner":        pet.Owner,
			"personality":  pet.Personality,
			"level":        pet.Level,
			"location":     pet.Location,
			"status":       pet.Status,
			"mood":         pet.Mood,
			"created_at":   pet.CreatedAt,
			"last_activity": pet.LastActivity,
		},
		"attributes": map[string]interface{}{
			"health":      pet.Health,
			"max_health":  pet.MaxHealth,
			"energy":      pet.Energy,
			"max_energy":  pet.MaxEnergy,
			"hunger":      pet.Hunger,
			"social":      pet.Social,
			"attack":      pet.Attack,
			"defense":     pet.Defense,
			"experience":  pet.Experience,
			"coins":       pet.Coins,
		},
		"social_data": map[string]interface{}{
			"friends": pet.Friends,
			"memory":  pet.Memory,
		},
		"capabilities": map[string]interface{}{
			"can_explore":   pet.CanExplore(),
			"can_rest":      pet.CanRest(),
			"can_socialize": pet.CanSocialize(),
			"is_alive":      pet.IsAlive(),
		},
	}

	return status, nil
}

func (ps *PetService) executeRestCommand(pet *models.Pet, params map[string]interface{}) (interface{}, error) {
	if !pet.CanRest() {
		return nil, fmt.Errorf("pet cannot rest at this time")
	}

	duration := 30
	if d, ok := params["duration"].(float64); ok {
		duration = int(d)
	}

	action := Action{
		Type:     ActionRest,
		Priority: 100,
		Reason:   "接受命令休息",
		Duration: duration,
	}

	ps.executeRestAction(pet, action)
	return map[string]interface{}{
		"action":   "rest",
		"duration": duration,
		"message":  fmt.Sprintf("%s 开始休息 %d 秒", pet.Name, duration),
	}, nil
}

func (ps *PetService) executeFeedCommand(pet *models.Pet, params map[string]interface{}) (interface{}, error) {
	amount := 20
	if a, ok := params["amount"].(float64); ok {
		amount = int(a)
	}

	cost := amount / 2
	if pet.Coins < cost {
		return nil, fmt.Errorf("not enough coins to feed pet")
	}

	pet.Coins -= cost
	pet.Feed(amount)

	event := models.Event{
		ID:        uuid.New().String(),
		PetID:     pet.ID,
		PetName:   pet.Name,
		Type:      models.EventReward,
		Message:   fmt.Sprintf("[%s] 通过命令喂食了%d点，消耗%d金币", pet.Name, amount, cost),
		Timestamp: time.Now(),
		Data:      models.EventData{Coins: -cost},
	}
	ps.addEvent(event)

	return map[string]interface{}{
		"action":   "feed",
		"amount":   amount,
		"cost":     cost,
		"message":  fmt.Sprintf("%s 进食了 %d 点", pet.Name, amount),
	}, nil
}

func (ps *PetService) executeSocializeCommand(pet *models.Pet, params map[string]interface{}) (interface{}, error) {
	if !pet.CanSocialize() {
		return nil, fmt.Errorf("pet cannot socialize at this time")
	}

	action := Action{
		Type:     ActionSocialize,
		Priority: 100,
		Reason:   "接受命令社交",
		Duration: 40,
	}

	ps.executeSocializeAction(pet, action)
	return map[string]interface{}{
		"action":  "socialize",
		"message": fmt.Sprintf("%s 开始社交", pet.Name),
	}, nil
}

func (ps *PetService) executeExploreCommand(pet *models.Pet, params map[string]interface{}) (interface{}, error) {
	if !pet.CanExplore() {
		return nil, fmt.Errorf("pet cannot explore at this time")
	}

	direction := "未知方向"
	if d, ok := params["direction"].(string); ok {
		direction = d
	}

	action := Action{
		Type:     ActionExplore,
		Priority: 100,
		Reason:   fmt.Sprintf("接受命令向%s探索", direction),
		Duration: 60,
	}

	ps.executeExploreAction(pet, action)
	return map[string]interface{}{
		"action":    "explore",
		"direction": direction,
		"message":   fmt.Sprintf("%s 开始向 %s 探索", pet.Name, direction),
	}, nil
}

func (ps *PetService) executeAddCoinsCommand(pet *models.Pet, params map[string]interface{}) (interface{}, error) {
	amount := 100
	if a, ok := params["amount"].(float64); ok {
		amount = int(a)
	}
	
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if amount > 10000 {
		return nil, fmt.Errorf("amount too large (max: 10000)")
	}
	
	oldCoins := pet.Coins
	pet.Coins += amount
	
	ps.addEvent(models.Event{
		ID:        uuid.New().String(),
		PetID:     pet.ID,
		PetName:   pet.Name,
		Type:      "debug",
		Message:   fmt.Sprintf("[%s] 调试指令：增加了 %d 金币 (总计: %d)", pet.Name, amount, pet.Coins),
		Timestamp: time.Now(),
		Data:      models.EventData{Coins: amount},
	})
	
	ps.savePetToDatabase(pet)
	
	return map[string]interface{}{
		"action":    "addcoins",
		"amount":    amount,
		"old_coins": oldCoins,
		"new_coins": pet.Coins,
		"message":   fmt.Sprintf("%s 获得了 %d 金币！当前总金币: %d", pet.Name, amount, pet.Coins),
	}, nil
}