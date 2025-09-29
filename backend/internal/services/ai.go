package services

import (
	"fmt"
	"time"

	"miningpet/internal/models"
	"github.com/google/uuid"
)

func (ps *PetService) runGlobalAI() {
	ticker := time.NewTicker(30 * time.Second)
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

func (ps *PetService) startPetAI(pet *models.Pet) {
	if _, exists := ps.activePets[pet.ID]; exists {
		return
	}

	ticker := time.NewTicker(15 * time.Second)
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

			if currentPet.Status != models.StatusIdle && currentPet.Status != "等待中" && currentPet.Status != models.StatusExploring {
				ps.mutex.Unlock()
				continue
			}

			action := ps.aiEngine.DecideNextAction(currentPet)
			ps.executeAction(currentPet, action)
			ps.mutex.Unlock()
		}
	}()
}

func (ps *PetService) updatePetAttributes(pet *models.Pet) {
	if pet.Energy > 0 {
		energyLoss := 2
		if pet.Status == models.StatusExploring || pet.Status == models.StatusFighting {
			energyLoss = 5
		}
		pet.ConsumeEnergy(energyLoss)
	}

	if pet.Hunger > 0 {
		hungerLoss := 3
		if pet.Status == models.StatusExploring {
			hungerLoss = 5
		}
		pet.ConsumeHunger(hungerLoss)
	}

	if pet.Social > 0 && pet.Status != models.StatusSocializing {
		pet.DecreaseSocial(1)
	}

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

	if pet.Status == "探索中" {
		return fmt.Errorf("pet is already exploring")
	}

	pet.Status = models.StatusExploring
	pet.LastActivity = time.Now()
	
	event := models.Event{
		ID:        uuid.New().String(),
		PetID:     pet.ID,
		PetName:   pet.Name,
		Type:      models.EventExplore,
		Message:   fmt.Sprintf("[%s] 收到探索指令，开始探索冒险...", pet.Name),
		Timestamp: time.Now(),
		Data:      models.EventData{Location: pet.Location},
	}
	ps.addEvent(event)
	
	go func() {
		time.Sleep(2 * time.Second)
		ps.mutex.Lock()
		if currentPet, exists := ps.pets[pet.ID]; exists && currentPet.Status == models.StatusExploring {
			action := ps.aiEngine.DecideNextAction(currentPet)
			ps.executeAction(currentPet, action)
		}
		ps.mutex.Unlock()
	}()
	
	return nil
}