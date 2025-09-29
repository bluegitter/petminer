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

type PetStatus string

const (
	StatusIdle      PetStatus = "等待中"
	StatusExploring PetStatus = "探索中"
	StatusFighting  PetStatus = "战斗中"
	StatusResting   PetStatus = "休息中"
	StatusSocializing PetStatus = "社交中"
)

type PetMood string

const (
	MoodHappy    PetMood = "开心"
	MoodNeutral  PetMood = "普通"
	MoodSad      PetMood = "沮丧"
	MoodExcited  PetMood = "兴奋"
	MoodTired    PetMood = "疲惫"
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
	Energy       int            `json:"energy"`        // 体力值 0-100
	MaxEnergy    int            `json:"max_energy"`
	Hunger       int            `json:"hunger"`        // 饱食度 0-100
	Social       int            `json:"social"`        // 社交度 0-100
	Mood         PetMood        `json:"mood"`          // 心情状态
	Attack       int            `json:"attack"`
	Defense      int            `json:"defense"`
	Coins        int            `json:"coins"`
	Location     string         `json:"location"`
	Status       PetStatus      `json:"status"`
	Memory       []string       `json:"memory"`        // 宠物记忆
	Friends      []string       `json:"friends"`       // 朋友列表
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

	pet := &Pet{
		ID:           uuid.New().String(),
		Name:         petNames[len(ownerName)%len(petNames)],
		Owner:        ownerName,
		Personality:  personalities[len(ownerName)%len(personalities)],
		Level:        1,
		Experience:   0,
		Health:       100,
		MaxHealth:    100,
		Energy:       100,
		MaxEnergy:    100,
		Hunger:       80,  // 稍微饿一点，需要关注
		Social:       50,  // 中等社交需求
		Mood:         MoodHappy,
		Attack:       10,
		Defense:      5,
		Coins:        0,
		Location:     "起始村庄",
		Status:       StatusIdle,
		Memory:       make([]string, 0),
		Friends:      make([]string, 0),
		LastActivity: time.Now(),
		CreatedAt:    time.Now(),
	}

	// 根据性格调整初始属性
	switch pet.Personality {
	case PersonalityBrave:
		pet.Attack += 3
		pet.Energy += 10
	case PersonalityGreedy:
		pet.Coins += 50
	case PersonalityFriendly:
		pet.Social += 20
	case PersonalityCautious:
		pet.Defense += 3
		pet.Health += 20
		pet.MaxHealth += 20
	case PersonalityCurious:
		pet.Experience += 20
	}

	return pet
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

// 状态管理方法
func (p *Pet) CanExplore() bool {
	return p.IsAlive() && p.Status == StatusIdle && p.Energy > 20
}

func (p *Pet) CanRest() bool {
	return p.IsAlive() && p.Status == StatusIdle
}

func (p *Pet) CanSocialize() bool {
	return p.IsAlive() && p.Status == StatusIdle && p.Social < 90
}

// 属性变化方法
func (p *Pet) ConsumeEnergy(amount int) {
	p.Energy -= amount
	if p.Energy < 0 {
		p.Energy = 0
	}
	p.updateMood()
}

func (p *Pet) RestoreEnergy(amount int) {
	p.Energy += amount
	if p.Energy > p.MaxEnergy {
		p.Energy = p.MaxEnergy
	}
	p.updateMood()
}

func (p *Pet) ConsumeHunger(amount int) {
	p.Hunger -= amount
	if p.Hunger < 0 {
		p.Hunger = 0
	}
	p.updateMood()
}

func (p *Pet) Feed(amount int) {
	p.Hunger += amount
	if p.Hunger > 100 {
		p.Hunger = 100
	}
	p.updateMood()
}

func (p *Pet) IncreaseSocial(amount int) {
	p.Social += amount
	if p.Social > 100 {
		p.Social = 100
	}
	p.updateMood()
}

func (p *Pet) DecreaseSocial(amount int) {
	p.Social -= amount
	if p.Social < 0 {
		p.Social = 0
	}
	p.updateMood()
}

// 记忆和朋友管理
func (p *Pet) AddMemory(memory string) {
	p.Memory = append(p.Memory, memory)
	if len(p.Memory) > 10 { // 只保留最近10条记忆
		p.Memory = p.Memory[1:]
	}
}

func (p *Pet) AddFriend(friendName string) {
	for _, friend := range p.Friends {
		if friend == friendName {
			return // 已经是朋友了
		}
	}
	p.Friends = append(p.Friends, friendName)
}

// 心情更新逻辑
func (p *Pet) updateMood() {
	score := 0
	
	// 健康状况影响
	healthPercent := float64(p.Health) / float64(p.MaxHealth)
	if healthPercent > 0.8 {
		score += 2
	} else if healthPercent < 0.3 {
		score -= 3
	}
	
	// 体力状况影响
	energyPercent := float64(p.Energy) / float64(p.MaxEnergy)
	if energyPercent > 0.8 {
		score += 1
	} else if energyPercent < 0.2 {
		score -= 2
	}
	
	// 饱食度影响
	if p.Hunger > 80 {
		score += 1
	} else if p.Hunger < 30 {
		score -= 2
	}
	
	// 社交度影响
	if p.Social > 70 {
		score += 1
	} else if p.Social < 30 {
		score -= 1
	}
	
	// 根据得分设置心情
	switch {
	case score >= 3:
		p.Mood = MoodExcited
	case score >= 1:
		p.Mood = MoodHappy
	case score >= -1:
		p.Mood = MoodNeutral
	case score >= -3:
		p.Mood = MoodSad
	default:
		p.Mood = MoodTired
	}
}

// 获取心情影响的行为倾向
func (p *Pet) GetMoodInfluence() float64 {
	switch p.Mood {
	case MoodExcited:
		return 1.5
	case MoodHappy:
		return 1.2
	case MoodNeutral:
		return 1.0
	case MoodSad:
		return 0.8
	case MoodTired:
		return 0.6
	default:
		return 1.0
	}
}