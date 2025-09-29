package services

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"miningpet/internal/models"
	"github.com/google/uuid"
)

func (ps *PetService) GetRecentEvents(limit int) []models.Event {
	dbEvents, err := ps.eventRepo.GetRecentEvents(limit)
	if err != nil {
		log.Printf("Warning: failed to get events from database: %v", err)
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
	
	events := make([]models.Event, len(dbEvents))
	for i, event := range dbEvents {
		events[i] = *event
	}
	
	return events
}

func (ps *PetService) GetEventChannel() <-chan models.Event {
	return ps.eventsCh
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
		discoveries := []string{"宝箱", "神秘水晶", "古老卷轴", "闪光宝石", "魔法药水", "远古符文", "珍稀矿石", "神秘遗物"}
		discovery := discoveries[rand.Intn(len(discoveries))]
		
		discoveryMessages := []string{
			"发现了%s，获得%d金币！",
			"在探索中找到%s，收获%d金币！", 
			"意外挖掘出%s，得到%d金币奖励！",
			"仔细搜索后发现%s，获得%d金币！",
			"幸运地遇到%s，赚得%d金币！",
		}
		messageTemplate := discoveryMessages[rand.Intn(len(discoveryMessages))]
		event.Message = fmt.Sprintf("[%s] %s", pet.Name, fmt.Sprintf(messageTemplate, discovery, coins))
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
			
			rewardMessages := []string{
				"找到了一些零散的金币：+%d",
				"发现了闪闪发光的硬币：+%d",
				"从地上捡到了金币：+%d", 
				"在岩石缝隙中发现金币：+%d",
				"挖出了埋在土里的金币：+%d",
				"在古老树根下找到金币：+%d",
			}
			messageTemplate := rewardMessages[rand.Intn(len(rewardMessages))]
			event.Message = fmt.Sprintf("[%s] %s", pet.Name, fmt.Sprintf(messageTemplate, coins))
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
	eventKey := fmt.Sprintf("%s:%s:%s", event.PetID, event.Type, event.Message)
	
	if event.Data.Coins != 0 {
		eventKey = fmt.Sprintf("%s:coins:%d", eventKey, event.Data.Coins)
	}
	
	if lastTime, exists := ps.recentEvents[eventKey]; exists {
		if time.Since(lastTime) < 15*time.Second {
			return
		}
	}
	
	ps.recentEvents[eventKey] = event.Timestamp
	
	for key, timestamp := range ps.recentEvents {
		if time.Since(timestamp) > 60*time.Second {
			delete(ps.recentEvents, key)
		}
	}
	
	if err := ps.eventRepo.CreateEvent(&event); err != nil {
		log.Printf("Warning: failed to save event to database: %v", err)
	}
	
	ps.events = append(ps.events, event)
	if len(ps.events) > 1000 {
		ps.events = ps.events[100:]
	}
	
	select {
	case ps.eventsCh <- event:
	default:
	}
}