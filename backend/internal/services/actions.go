package services

import (
	"fmt"
	"math/rand"
	"time"

	"miningpet/internal/models"
	"github.com/google/uuid"
)

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
	}
}

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

	go func() {
		time.Sleep(time.Duration(action.Duration) * time.Second)
		ps.mutex.Lock()
		if pet.Status == models.StatusExploring {
			ps.processExploreResult(pet)
		}
		ps.mutex.Unlock()
	}()
}

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
			
			ps.savePetToDatabase(pet)
		}
		ps.mutex.Unlock()
	}()
}

func (ps *PetService) executeSocializeAction(pet *models.Pet, action Action) {
	pet.Status = models.StatusSocializing
	
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
			
			ps.savePetToDatabase(pet)
		}
		ps.mutex.Unlock()
	}()
}

func (ps *PetService) executeEatAction(pet *models.Pet, action Action) {
	if pet.Coins < 10 {
		pet.Status = "寻找食物"
		
		event := models.Event{
			ID:        uuid.New().String(),
			PetID:     pet.ID,
			PetName:   pet.Name,
			Type:      models.EventReward,
			Message:   fmt.Sprintf("[%s] 寻找免费的食物...", pet.Name),
			Timestamp: time.Now(),
		}
		ps.addEvent(event)
		
		go func() {
			time.Sleep(time.Duration(action.Duration) * time.Second)
			ps.mutex.Lock()
			if pet.Status == "寻找食物" {
				feedAmount := 10 + rand.Intn(15)
				pet.Feed(feedAmount)
				pet.Status = models.StatusIdle
				
				foundEvent := models.Event{
					ID:        uuid.New().String(),
					PetID:     pet.ID,
					PetName:   pet.Name,
					Type:      models.EventReward,
					Message:   fmt.Sprintf("[%s] 找到了一些免费食物，饱食度+%d", pet.Name, feedAmount),
					Timestamp: time.Now(),
				}
				ps.addEvent(foundEvent)
				
				ps.savePetToDatabase(pet)
			}
			ps.mutex.Unlock()
		}()
	} else {
		cost := 10 + rand.Intn(10)
		if pet.Coins >= cost {
			pet.Status = "进食中"
			pet.Coins -= cost
			
			event := models.Event{
				ID:        uuid.New().String(),
				PetID:     pet.ID,
				PetName:   pet.Name,
				Type:      models.EventReward,
				Message:   fmt.Sprintf("[%s] 购买食物中...", pet.Name),
				Timestamp: time.Now(),
			}
			ps.addEvent(event)
			
			go func() {
				time.Sleep(time.Duration(action.Duration) * time.Second)
				ps.mutex.Lock()
				if pet.Status == "进食中" {
					feedAmount := 25 + rand.Intn(20)
					pet.Feed(feedAmount)
					pet.Status = models.StatusIdle
					
					finishEvent := models.Event{
						ID:        uuid.New().String(),
						PetID:     pet.ID,
						PetName:   pet.Name,
						Type:      models.EventReward,
						Message:   fmt.Sprintf("[%s] 花费%d金币买了美味的食物，饱食度+%d", pet.Name, cost, feedAmount),
						Timestamp: time.Now(),
						Data:      models.EventData{Coins: -cost},
					}
					ps.addEvent(finishEvent)
					
					ps.savePetToDatabase(pet)
				}
				ps.mutex.Unlock()
			}()
		}
	}
}

func (ps *PetService) processExploreResult(pet *models.Pet) {
	event := ps.generateRandomEvent(pet)
	ps.addEvent(event)
	pet.Status = models.StatusIdle
	pet.LastActivity = time.Now()
	
	ps.savePetToDatabase(pet)
}