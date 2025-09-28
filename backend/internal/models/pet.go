package models

import (
	"time"

	"github.com/google/uuid"
)

type PetPersonality string

const (
	PersonalityBrave    PetPersonality = "brave"
	PersonalityGreedy   PetPersonality = "greedy"
	PersonalityFriendly PetPersonality = "friendly"
	PersonalityCautious PetPersonality = "cautious"
	PersonalityCurious  PetPersonality = "curious"
)

type Pet struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Owner        string         `json:"owner"`
	Personality  PetPersonality `json:"personality"`
	Level        int            `json:"level"`
	Experience   int            `json:"experience"`
	Health       int            `json:"health"`
	MaxHealth    int            `json:"max_health"`
	Attack       int            `json:"attack"`
	Defense      int            `json:"defense"`
	Coins        int            `json:"coins"`
	Location     string         `json:"location"`
	Status       string         `json:"status"`
	LastActivity time.Time      `json:"last_activity"`
	CreatedAt    time.Time      `json:"created_at"`
}

type Item struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Rarity   string `json:"rarity"`
	Value    int    `json:"value"`
	Quantity int    `json:"quantity"`
}

type Inventory struct {
	PetID string `json:"pet_id"`
	Items []Item `json:"items"`
}

func NewPet(ownerName string) *Pet {
	personalities := []PetPersonality{
		PersonalityBrave, PersonalityGreedy, PersonalityFriendly,
		PersonalityCautious, PersonalityCurious,
	}
	
	petNames := []string{
		"Lucky", "Brave", "Shadow", "Spark", "Whisper",
		"Thunder", "Frost", "Blaze", "Swift", "Mystic",
	}

	return &Pet{
		ID:           uuid.New().String(),
		Name:         petNames[len(ownerName)%len(petNames)],
		Owner:        ownerName,
		Personality:  personalities[len(ownerName)%len(personalities)],
		Level:        1,
		Experience:   0,
		Health:       100,
		MaxHealth:    100,
		Attack:       10,
		Defense:      5,
		Coins:        0,
		Location:     "起始村庄",
		Status:       "等待中",
		LastActivity: time.Now(),
		CreatedAt:    time.Now(),
	}
}

func (p *Pet) GainExperience(exp int) bool {
	p.Experience += exp
	if p.Experience >= p.Level*100 {
		p.LevelUp()
		return true
	}
	return false
}

func (p *Pet) LevelUp() {
	p.Level++
	p.MaxHealth += 20
	p.Health = p.MaxHealth
	p.Attack += 5
	p.Defense += 3
	p.Experience = 0
}

func (p *Pet) TakeDamage(damage int) {
	actualDamage := damage - p.Defense
	if actualDamage < 1 {
		actualDamage = 1
	}
	p.Health -= actualDamage
	if p.Health < 0 {
		p.Health = 0
	}
}

func (p *Pet) Heal(amount int) {
	p.Health += amount
	if p.Health > p.MaxHealth {
		p.Health = p.MaxHealth
	}
}

func (p *Pet) IsAlive() bool {
	return p.Health > 0
}