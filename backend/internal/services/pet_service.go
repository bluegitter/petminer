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
}

func NewPetService() *PetService {
	return &PetService{
		pets:     make(map[string]*models.Pet),
		events:   make([]models.Event, 0),
		mutex:    sync.RWMutex{},
		eventsCh: make(chan models.Event, 100),
	}
}

func (ps *PetService) CreatePet(ownerName string) *models.Pet {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	pet := models.NewPet(ownerName)
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
	return pet
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
	
	return (petPower + personalityBonus + rand.Intn(20)) > (monsterPower + rand.Intn(15))
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