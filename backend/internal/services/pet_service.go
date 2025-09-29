package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"miningpet/internal/database"
	"miningpet/internal/models"
	"github.com/google/uuid"
)

type PetService struct {
	pets         map[string]*models.Pet
	events       []models.Event
	mutex        sync.RWMutex
	eventsCh     chan models.Event
	aiEngine     *AIEngine
	activePets   map[string]*time.Ticker
	recentEvents map[string]time.Time
	
	petRepo   *database.PetRepository
	eventRepo *database.EventRepository
}

func NewPetService() *PetService {
	ps := &PetService{
		pets:         make(map[string]*models.Pet),
		events:       make([]models.Event, 0),
		mutex:        sync.RWMutex{},
		eventsCh:     make(chan models.Event, 100),
		aiEngine:     NewAIEngine(),
		activePets:   make(map[string]*time.Ticker),
		recentEvents: make(map[string]time.Time),
		petRepo:      database.NewPetRepository(),
		eventRepo:    database.NewEventRepository(),
	}
	
	if err := ps.loadPetsFromDatabase(); err != nil {
		log.Printf("Warning: failed to load pets from database: %v", err)
	}
	
	go ps.runGlobalAI()
	ps.startExistingPetsAI()
	
	return ps
}

func (ps *PetService) loadPetsFromDatabase() error {
	pets, err := ps.petRepo.GetAllPets()
	if err != nil {
		return fmt.Errorf("failed to load pets from database: %w", err)
	}

	for _, pet := range pets {
		ps.pets[pet.ID] = pet
	}

	log.Printf("Loaded %d pets from database", len(pets))
	return nil
}

func (ps *PetService) startExistingPetsAI() {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	for _, pet := range ps.pets {
		if pet.IsAlive() {
			ps.startPetAI(pet)
		}
	}
}

func (ps *PetService) savePetToDatabase(pet *models.Pet) {
	go func() {
		if err := ps.petRepo.UpdatePet(pet); err != nil {
			log.Printf("Warning: failed to save pet %s to database: %v", pet.ID, err)
		}
	}()
}

func (ps *PetService) CreatePet(ownerName string) (*models.Pet, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	for _, existingPet := range ps.pets {
		if existingPet.Owner == ownerName {
			return nil, fmt.Errorf("用户 %s 已经拥有宠物 %s，每位训练师只能拥有一只宠物", ownerName, existingPet.Name)
		}
	}

	if existingPet, err := ps.petRepo.GetPetByOwner(ownerName); err != nil {
		log.Printf("Error checking existing pet in database: %v", err)
	} else if existingPet != nil {
		return nil, fmt.Errorf("用户 %s 已经拥有宠物 %s，每位训练师只能拥有一只宠物", ownerName, existingPet.Name)
	}

	pet := models.NewPet(ownerName)
	
	if err := ps.petRepo.CreatePet(pet); err != nil {
		return nil, fmt.Errorf("failed to save pet to database: %w", err)
	}
	
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
	ps.startPetAI(pet)
	
	return pet, nil
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

func (ps *PetService) GetPetByOwner(ownerName string) (*models.Pet, error) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()
	
	for _, pet := range ps.pets {
		if pet.Owner == ownerName {
			return pet, nil
		}
	}
	
	return ps.petRepo.GetPetByOwner(ownerName)
}