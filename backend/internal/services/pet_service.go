package services

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"miningpet/internal/models"
	"github.com/google/uuid"
)

type PetService struct {
	pets     map[string]*models.Pet
	events   []models.Event
	mutex    sync.RWMutex
	eventsCh chan models.Event
	aiEngine *AIEngine
	activePets map[string]*time.Ticker // 活跃的宠物和它们的ticker
}

func NewPetService() *PetService {
	ps := &PetService{
		pets:       make(map[string]*models.Pet),
		events:     make([]models.Event, 0),
		mutex:      sync.RWMutex{},
		eventsCh:   make(chan models.Event, 100),
		aiEngine:   NewAIEngine(),
		activePets: make(map[string]*time.Ticker),
	}
	
	// 启动全局AI循环
	go ps.runGlobalAI()
	
	return ps
}

func (ps *PetService) CreatePet(ownerName string) (*models.Pet, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	// 检查该用户是否已经有宠物
	for _, existingPet := range ps.pets {
		if existingPet.Owner == ownerName {
			return nil, fmt.Errorf("用户 %s 已经拥有宠物 %s，每位训练师只能拥有一只宠物", ownerName, existingPet.Name)
		}
	}

	pet := models.NewPet(ownerName)
	ps.pets[pet.ID] = pet

	event := models.Event{
		ID:        uuid.New().String(),
		PetID:     pet.ID,
		PetName:   pet.Name,
		Type:      models.EventExplore,
		Message:   fmt.Sprintf("[%s] 诞生了！点击骰子选择种族和技能后开始探索世界...", pet.Name),
		Timestamp: time.Now(),
		Data:      models.EventData{Location: pet.Location},
	}

	ps.addEvent(event)

	return pet, nil
}

// RollRace 掷骰子选择种族
func (ps *PetService) RollRace(petID string) (*models.RaceInfo, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	pet, exists := ps.pets[petID]
	if !exists {
		return nil, fmt.Errorf("pet not found")
	}

	// 检查是否已经有种族
	if pet.Race.Name != "" {
		return nil, fmt.Errorf("pet already has a race")
	}

	race := models.GenerateRandomRace()
	pet.Race = race
	pet.ApplyRaceBonuses()

	event := models.Event{
		ID:        uuid.New().String(),
		PetID:     pet.ID,
		PetName:   pet.Name,
		Type:      models.EventReward,
		Message:   fmt.Sprintf("🎲 [%s] 掷出了种族：%s (%s) - %s品质！", pet.Name, race.Name, race.Category, race.Rarity),
		Timestamp: time.Now(),
		Data:      models.EventData{},
	}
	ps.addEvent(event)

	return &race, nil
}

// RollSkill 掷骰子选择技能
func (ps *PetService) RollSkill(petID string) (*models.PetSkill, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	pet, exists := ps.pets[petID]
	if !exists {
		return nil, fmt.Errorf("pet not found")
	}

	// 检查是否已经有技能
	if pet.Skill.Name != "" {
		return nil, fmt.Errorf("pet already has a skill")
	}

	skill := models.GenerateRandomSkill()
	pet.Skill = skill

	event := models.Event{
		ID:        uuid.New().String(),
		PetID:     pet.ID,
		PetName:   pet.Name,
		Type:      models.EventReward,
		Message:   fmt.Sprintf("🎲 [%s] 掷出了技能：%s (%s) - %s品质！", pet.Name, skill.Name, skill.Type, skill.Rarity),
		Timestamp: time.Now(),
		Data:      models.EventData{},
	}
	ps.addEvent(event)

	// 如果种族和技能都选择完毕，启动AI
	if pet.Race.Name != "" && pet.Skill.Name != "" {
		ps.startPetAI(pet)

		finalEvent := models.Event{
			ID:        uuid.New().String(),
			PetID:     pet.ID,
			PetName:   pet.Name,
			Type:      models.EventExplore,
			Message:   fmt.Sprintf("✨ [%s] 完成初始化，开始探索世界！", pet.Name),
			Timestamp: time.Now(),
			Data:      models.EventData{Location: pet.Location},
		}
		ps.addEvent(finalEvent)
	}

	return &skill, nil
}

func (ps *PetService) GetPet(petID string) (*models.Pet, bool) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	
	pet, exists := ps.pets[petID]
	return pet, exists
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

func (ps *PetService) GetRecentEvents(limit int) []models.Event {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	
	if limit <= 0 || limit > len(ps.events) {
		limit = len(ps.events)
	}
	
	start := len(ps.events) - limit
	if start < 0 {
		start = 0
	}
	
	return ps.events[start:]
}

func (ps *PetService) GetEventChannel() <-chan models.Event {
	return ps.eventsCh
}

func (ps *PetService) StartExploration(petID string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	pet, exists := ps.pets[petID]
	if !exists {
		return fmt.Errorf("pet not found")
	}

	if !pet.IsAlive() {
		return fmt.Errorf("pet is not alive")
	}

	pet.Status = "探索中"
	pet.LastActivity = time.Now()
	
	go ps.exploreLoop(pet)
	return nil
}

func (ps *PetService) exploreLoop(pet *models.Pet) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ps.mutex.Lock()
		if pet.Status != "探索中" || !pet.IsAlive() {
			ps.mutex.Unlock()
			break
		}
		
		event := ps.generateRandomEvent(pet)
		ps.addEvent(event)
		pet.LastActivity = time.Now()
		ps.mutex.Unlock()
	}
}

func (ps *PetService) generateRandomEvent(pet *models.Pet) models.Event {
	eventTypes := []models.EventType{
		models.EventExplore, models.EventBattle, models.EventDiscovery,
		models.EventSocial, models.EventReward,
	}

	eventType := eventTypes[rand.Intn(len(eventTypes))]
	event := models.Event{
		ID:        uuid.New().String(),
		PetID:     pet.ID,
		PetName:   pet.Name,
		Type:      eventType,
		Timestamp: time.Now(),
	}

	switch eventType {
	case models.EventExplore:
		location := models.Locations[rand.Intn(len(models.Locations))]
		pet.Location = location
		event.Message = fmt.Sprintf("[%s] 来到了%s，开始探索...", pet.Name, location)
		event.Data.Location = location

	case models.EventBattle:
		monster := models.Monsters[rand.Intn(len(models.Monsters))]
		victory := ps.simulateBattle(pet, monster)
		
		if victory {
			pet.GainExperience(monster.ExpReward)
			pet.Coins += monster.CoinReward
			event.Message = fmt.Sprintf("[%s] 击败了%s！获得经验+%d，金币+%d", 
				pet.Name, monster.Name, monster.ExpReward, monster.CoinReward)
		} else {
			damage := monster.Attack - pet.Defense
			if damage < 1 {
				damage = 1
			}
			pet.TakeDamage(damage)
			event.Message = fmt.Sprintf("[%s] 被%s击败，受到%d点伤害", pet.Name, monster.Name, damage)
		}
		
		event.Data.Enemy = monster.Name
		event.Data.IsVictory = victory
		event.Data.Experience = monster.ExpReward
		event.Data.Coins = monster.CoinReward

	case models.EventDiscovery:
		coins := rand.Intn(20) + 5
		pet.Coins += coins
		discoveries := []string{"宝箱", "神秘水晶", "古老卷轴", "闪光宝石", "魔法药水"}
		discovery := discoveries[rand.Intn(len(discoveries))]
		event.Message = fmt.Sprintf("[%s] 发现了%s，获得%d金币！", pet.Name, discovery, coins)
		event.Data.Coins = coins

	case models.EventSocial:
		friends := []string{"小明", "小红", "阿强", "丽丽", "小虎"}
		friend := friends[rand.Intn(len(friends))]
		event.Message = fmt.Sprintf("[%s] 遇到了%s的宠物，成为了朋友！", pet.Name, friend)
		event.Data.FriendName = friend

	case models.EventReward:
		if rand.Intn(100) < 5 {
			event.Type = models.EventRareFind
			rareReward := rand.Intn(1000) + 500
			pet.Coins += rareReward
			event.Message = fmt.Sprintf("[%s] 🌟 发现神秘矿石！获得大奖%d金币！", pet.Name, rareReward)
			event.Data.Coins = rareReward
		} else {
			coins := rand.Intn(50) + 10
			pet.Coins += coins
			event.Message = fmt.Sprintf("[%s] 找到了一些零散的金币：+%d", pet.Name, coins)
			event.Data.Coins = coins
		}
	}

	return event
}

func (ps *PetService) simulateBattle(pet *models.Pet, monster models.Monster) bool {
	petPower := pet.Attack + pet.Defense + pet.Level*2
	monsterPower := monster.Attack + monster.Defense

	personalityBonus := 0
	switch pet.Personality {
	case models.PersonalityBrave:
		personalityBonus = 5
	case models.PersonalityGreedy:
		personalityBonus = 2
	case models.PersonalityCautious:
		personalityBonus = 3
	}

	// 技能效果
	skillBonus := 0
	if pet.Skill.Name != "" {
		switch pet.Skill.Type {
		case models.SkillTypeAttack:
			// 攻击技能增加战斗力
			skillBonus = int(pet.Skill.Level) * 3
		case models.SkillTypeDefense:
			// 防御技能增加防御力
			skillBonus = int(pet.Skill.Level) * 2
		case models.SkillTypeVampire:
			// 吸血技能在胜利时回复生命值
			skillBonus = int(pet.Skill.Level) * 2
		}
	}

	victory := (petPower + personalityBonus + skillBonus + rand.Intn(20)) > (monsterPower + rand.Intn(15))

	// 技能特殊效果
	if victory && pet.Skill.Type == models.SkillTypeVampire {
		healAmount := int(pet.Skill.Level) * 5
		pet.Heal(healAmount)
	}

	return victory
}

func (ps *PetService) addEvent(event models.Event) {
	ps.events = append(ps.events, event)
	if len(ps.events) > 1000 {
		ps.events = ps.events[100:]
	}
	
	select {
	case ps.eventsCh <- event:
	default:
	}
}

// runGlobalAI 运行全局AI循环，管理宠物的属性衰减
func (ps *PetService) runGlobalAI() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒执行一次
	defer ticker.Stop()

	for range ticker.C {
		ps.mutex.Lock()
		for _, pet := range ps.pets {
			if pet.IsAlive() {
				ps.updatePetAttributes(pet)
			}
		}
		ps.mutex.Unlock()
	}
}

// startPetAI 启动单个宠物的AI循环
func (ps *PetService) startPetAI(pet *models.Pet) {
	if _, exists := ps.activePets[pet.ID]; exists {
		return // 已经启动了
	}

	ticker := time.NewTicker(15 * time.Second) // 每15秒决策一次
	ps.activePets[pet.ID] = ticker

	go func() {
		defer ticker.Stop()
		for range ticker.C {
			ps.mutex.Lock()
			currentPet, exists := ps.pets[pet.ID]
			if !exists || !currentPet.IsAlive() {
				delete(ps.activePets, pet.ID)
				ps.mutex.Unlock()
				return
			}

			// 如果宠物在忙碌状态，跳过此次决策
			if currentPet.Status != models.StatusIdle {
				ps.mutex.Unlock()
				continue
			}

			// AI决策
			action := ps.aiEngine.DecideNextAction(currentPet)
			ps.executeAction(currentPet, action)
			ps.mutex.Unlock()
		}
	}()
}

// updatePetAttributes 更新宠物属性（自然衰减）
func (ps *PetService) updatePetAttributes(pet *models.Pet) {
	// 体力自然衰减
	if pet.Energy > 0 {
		energyLoss := 2
		if pet.Status == models.StatusExploring || pet.Status == models.StatusFighting {
			energyLoss = 5
		}
		pet.ConsumeEnergy(energyLoss)
	}

	// 饱食度自然衰减
	if pet.Hunger > 0 {
		hungerLoss := 3
		if pet.Status == models.StatusExploring {
			hungerLoss = 5
		}
		pet.ConsumeHunger(hungerLoss)
	}

	// 社交度缓慢衰减
	if pet.Social > 0 && pet.Status != models.StatusSocializing {
		pet.DecreaseSocial(1)
	}

	// 如果饱食度太低，影响健康
	if pet.Hunger < 20 && pet.Health > 0 {
		pet.TakeDamage(5)
		ps.addEvent(models.Event{
			ID:        uuid.New().String(),
			PetID:     pet.ID,
			PetName:   pet.Name,
			Type:      models.EventReward,
			Message:   fmt.Sprintf("[%s] 因为饥饿失去了5点生命值", pet.Name),
			Timestamp: time.Now(),
			Data:      models.EventData{Damage: 5},
		})
	}
}

// executeAction 执行AI决定的行为
func (ps *PetService) executeAction(pet *models.Pet, action Action) {
	switch action.Type {
	case ActionExplore:
		ps.executeExploreAction(pet, action)
	case ActionRest:
		ps.executeRestAction(pet, action)
	case ActionSocialize:
		ps.executeSocializeAction(pet, action)
	case ActionEat:
		ps.executeEatAction(pet, action)
	case ActionIdle:
		// 空闲状态，不需要特殊处理
	}
}

// executeExploreAction 执行探索行为
func (ps *PetService) executeExploreAction(pet *models.Pet, action Action) {
	pet.Status = models.StatusExploring
	
	event := models.Event{
		ID:        uuid.New().String(),
		PetID:     pet.ID,
		PetName:   pet.Name,
		Type:      models.EventExplore,
		Message:   fmt.Sprintf("[%s] %s", pet.Name, action.Reason),
		Timestamp: time.Now(),
		Data:      models.EventData{Location: pet.Location},
	}
	ps.addEvent(event)

	// 延迟执行探索结果
	go func() {
		time.Sleep(time.Duration(action.Duration) * time.Second)
		ps.mutex.Lock()
		if pet.Status == models.StatusExploring {
			ps.processExploreResult(pet)
		}
		ps.mutex.Unlock()
	}()
}

// executeRestAction 执行休息行为
func (ps *PetService) executeRestAction(pet *models.Pet, action Action) {
	pet.Status = models.StatusResting
	
	event := models.Event{
		ID:        uuid.New().String(),
		PetID:     pet.ID,
		PetName:   pet.Name,
		Type:      models.EventReward,
		Message:   fmt.Sprintf("[%s] %s", pet.Name, action.Reason),
		Timestamp: time.Now(),
	}
	ps.addEvent(event)

	go func() {
		time.Sleep(time.Duration(action.Duration) * time.Second)
		ps.mutex.Lock()
		if pet.Status == models.StatusResting {
			restoreAmount := 20 + rand.Intn(20)
			pet.RestoreEnergy(restoreAmount)
			pet.Heal(10)
			pet.Status = models.StatusIdle
			
			ps.addEvent(models.Event{
				ID:        uuid.New().String(),
				PetID:     pet.ID,
				PetName:   pet.Name,
				Type:      models.EventReward,
				Message:   fmt.Sprintf("[%s] 休息完毕，恢复了%d点体力", pet.Name, restoreAmount),
				Timestamp: time.Now(),
			})
		}
		ps.mutex.Unlock()
	}()
}

// executeSocializeAction 执行社交行为
func (ps *PetService) executeSocializeAction(pet *models.Pet, action Action) {
	pet.Status = models.StatusSocializing
	
	// 寻找其他宠物进行社交
	var socialPartner *models.Pet
	for _, otherPet := range ps.pets {
		if otherPet.ID != pet.ID && otherPet.IsAlive() {
			socialPartner = otherPet
			break
		}
	}
	
	var message string
	if socialPartner != nil {
		message = fmt.Sprintf("[%s] 与 %s 愉快地交流", pet.Name, socialPartner.Name)
		pet.AddFriend(socialPartner.Owner)
		socialPartner.AddFriend(pet.Owner)
	} else {
		message = fmt.Sprintf("[%s] %s", pet.Name, action.Reason)
	}
	
	event := models.Event{
		ID:        uuid.New().String(),
		PetID:     pet.ID,
		PetName:   pet.Name,
		Type:      models.EventSocial,
		Message:   message,
		Timestamp: time.Now(),
	}
	ps.addEvent(event)

	go func() {
		time.Sleep(time.Duration(action.Duration) * time.Second)
		ps.mutex.Lock()
		if pet.Status == models.StatusSocializing {
			socialGain := 15 + rand.Intn(20)
			pet.IncreaseSocial(socialGain)
			pet.Status = models.StatusIdle
			
			ps.addEvent(models.Event{
				ID:        uuid.New().String(),
				PetID:     pet.ID,
				PetName:   pet.Name,
				Type:      models.EventSocial,
				Message:   fmt.Sprintf("[%s] 社交结束，心情变好了", pet.Name),
				Timestamp: time.Now(),
			})
		}
		ps.mutex.Unlock()
	}()
}

// executeEatAction 执行进食行为
func (ps *PetService) executeEatAction(pet *models.Pet, action Action) {
	if pet.Coins < 10 {
		// 没钱买食物，寻找免费食物
		event := models.Event{
			ID:        uuid.New().String(),
			PetID:     pet.ID,
			PetName:   pet.Name,
			Type:      models.EventReward,
			Message:   fmt.Sprintf("[%s] 寻找免费的食物...", pet.Name),
			Timestamp: time.Now(),
		}
		ps.addEvent(event)
		
		feedAmount := 10 + rand.Intn(15)
		pet.Feed(feedAmount)
	} else {
		// 花钱买食物
		cost := 10 + rand.Intn(10)
		if pet.Coins >= cost {
			pet.Coins -= cost
			feedAmount := 25 + rand.Intn(20)
			pet.Feed(feedAmount)
			
			event := models.Event{
				ID:        uuid.New().String(),
				PetID:     pet.ID,
				PetName:   pet.Name,
				Type:      models.EventReward,
				Message:   fmt.Sprintf("[%s] 花费%d金币买了美味的食物，饱食度+%d", pet.Name, cost, feedAmount),
				Timestamp: time.Now(),
				Data:      models.EventData{Coins: -cost},
			}
			ps.addEvent(event)
		}
	}
}

// processExploreResult 处理探索结果（复用原有逻辑）
func (ps *PetService) processExploreResult(pet *models.Pet) {
	event := ps.generateRandomEvent(pet)
	ps.addEvent(event)
	pet.Status = models.StatusIdle
	pet.LastActivity = time.Now()
}

// ExecuteCommand 通用命令执行接口
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
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// RestPet 让宠物休息
func (ps *PetService) RestPet(petID string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	pet, exists := ps.pets[petID]
	if !exists {
		return fmt.Errorf("pet not found")
	}

	if !pet.CanRest() {
		return fmt.Errorf("pet cannot rest at this time")
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

// FeedPet 给宠物喂食
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

	// 检查是否有足够的金币
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

// SocializePet 让宠物社交
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

// GetPetStatus 获取宠物详细状态
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
		"race_info": map[string]interface{}{
			"name":          pet.Race.Name,
			"category":      pet.Race.Category,
			"rarity":        pet.Race.Rarity,
			"health_bonus":  pet.Race.HealthBonus,
			"attack_bonus":  pet.Race.AttackBonus,
			"defense_bonus": pet.Race.DefenseBonus,
		},
		"skill_info": map[string]interface{}{
			"type":   pet.Skill.Type,
			"level":  pet.Skill.Level,
			"name":   pet.Skill.Name,
			"rarity": pet.Skill.Rarity,
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

// 命令执行的具体实现
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