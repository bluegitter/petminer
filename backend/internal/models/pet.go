package models

import (
	"math/rand"
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

// 宠物种族类型
type PetRace string

const (
	// 血牛类 (生命加成)
	RaceElephant PetRace = "大象"
	RaceBear     PetRace = "熊"
	RaceMoose    PetRace = "驼鹿"

	// 攻击类 (攻击力加成)
	RaceTiger   PetRace = "老虎"
	RaceLion    PetRace = "狮子"
	RaceHyena   PetRace = "鬣狗"
	RaceWolf    PetRace = "狼"
	RaceLeopard PetRace = "豹子"

	// 防御类 (防御力加成)
	RaceRhino      PetRace = "犀牛"
	RaceTurtle     PetRace = "乌龟"
	RacePangolin   PetRace = "穿山甲"
)

// 技能类型
type SkillType string

const (
	SkillTypeAttack   SkillType = "攻击"
	SkillTypeDefense  SkillType = "防御"
	SkillTypeVampire  SkillType = "吸血"
)

// 技能等级
type SkillLevel int

const (
	SkillLevel1 SkillLevel = 1
	SkillLevel2 SkillLevel = 2
	SkillLevel3 SkillLevel = 3
)

// 技能定义
type PetSkill struct {
	Type   SkillType  `json:"type"`
	Level  SkillLevel `json:"level"`
	Name   string     `json:"name"`
	Rarity string     `json:"rarity"`
}

// 种族定义
type RaceInfo struct {
	Name        PetRace `json:"name"`
	Category    string  `json:"category"`
	Rarity      string  `json:"rarity"`
	HealthBonus int     `json:"health_bonus"`
	AttackBonus int     `json:"attack_bonus"`
	DefenseBonus int    `json:"defense_bonus"`
}

type Pet struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Owner        string         `json:"owner"`
	Personality  PetPersonality `json:"personality"`
	Race         RaceInfo       `json:"race"`          // 种族信息
	Skill        PetSkill       `json:"skill"`         // 技能信息
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

// 种族配置数据
var RaceConfigs = map[PetRace]RaceInfo{
	// 血牛类 (生命加成)
	RaceElephant: {Name: RaceElephant, Category: "血牛", Rarity: "紫色", HealthBonus: 60, AttackBonus: 0, DefenseBonus: 0},
	RaceBear:     {Name: RaceBear, Category: "血牛", Rarity: "蓝色", HealthBonus: 40, AttackBonus: 0, DefenseBonus: 0},
	RaceMoose:    {Name: RaceMoose, Category: "血牛", Rarity: "绿色", HealthBonus: 20, AttackBonus: 0, DefenseBonus: 0},

	// 攻击类 (攻击力加成)
	RaceTiger:   {Name: RaceTiger, Category: "攻击", Rarity: "紫色", HealthBonus: 0, AttackBonus: 12, DefenseBonus: 0},
	RaceLion:    {Name: RaceLion, Category: "攻击", Rarity: "蓝色", HealthBonus: 0, AttackBonus: 8, DefenseBonus: 0},
	RaceHyena:   {Name: RaceHyena, Category: "攻击", Rarity: "蓝色", HealthBonus: 0, AttackBonus: 6, DefenseBonus: 0},
	RaceWolf:    {Name: RaceWolf, Category: "攻击", Rarity: "绿色", HealthBonus: 0, AttackBonus: 5, DefenseBonus: 0},
	RaceLeopard: {Name: RaceLeopard, Category: "攻击", Rarity: "绿色", HealthBonus: 0, AttackBonus: 3, DefenseBonus: 0},

	// 防御类 (防御力加成)
	RaceRhino:    {Name: RaceRhino, Category: "防御", Rarity: "紫色", HealthBonus: 0, AttackBonus: 0, DefenseBonus: 8},
	RaceTurtle:   {Name: RaceTurtle, Category: "防御", Rarity: "蓝色", HealthBonus: 0, AttackBonus: 0, DefenseBonus: 6},
	RacePangolin: {Name: RacePangolin, Category: "防御", Rarity: "绿色", HealthBonus: 0, AttackBonus: 0, DefenseBonus: 4},
}

// 技能配置数据
var SkillConfigs = map[SkillType]map[SkillLevel]PetSkill{
	SkillTypeAttack: {
		SkillLevel1: {Type: SkillTypeAttack, Level: SkillLevel1, Name: "爪击", Rarity: "绿色"},
		SkillLevel2: {Type: SkillTypeAttack, Level: SkillLevel2, Name: "撕咬", Rarity: "蓝色"},
		SkillLevel3: {Type: SkillTypeAttack, Level: SkillLevel3, Name: "突袭", Rarity: "紫色"},
	},
	SkillTypeDefense: {
		SkillLevel1: {Type: SkillTypeDefense, Level: SkillLevel1, Name: "反弹", Rarity: "绿色"},
		SkillLevel2: {Type: SkillTypeDefense, Level: SkillLevel2, Name: "铁刺", Rarity: "蓝色"},
		SkillLevel3: {Type: SkillTypeDefense, Level: SkillLevel3, Name: "钢针", Rarity: "紫色"},
	},
	SkillTypeVampire: {
		SkillLevel1: {Type: SkillTypeVampire, Level: SkillLevel1, Name: "蝙蝠之咬", Rarity: "绿色"},
		SkillLevel2: {Type: SkillTypeVampire, Level: SkillLevel2, Name: "狼人之咬", Rarity: "蓝色"},
		SkillLevel3: {Type: SkillTypeVampire, Level: SkillLevel3, Name: "吸血鬼之咬", Rarity: "紫色"},
	},
}

// 种族出现概率配置 (可动态调节)
var RaceWeights = map[PetRace]int{
	// 血牛类
	RaceElephant: 5,  // 紫色 5%
	RaceBear:     15, // 蓝色 15%
	RaceMoose:    30, // 绿色 30%

	// 攻击类
	RaceTiger:   5,  // 紫色 5%
	RaceLion:    10, // 蓝色 10%
	RaceHyena:   10, // 蓝色 10%
	RaceWolf:    15, // 绿色 15%
	RaceLeopard: 15, // 绿色 15%

	// 防御类
	RaceRhino:    5,  // 紫色 5%
	RaceTurtle:   10, // 蓝色 10%
	RacePangolin: 15, // 绿色 15%
}

// 技能出现概率配置 (可动态调节)
var SkillWeights = map[SkillType]map[SkillLevel]int{
	SkillTypeAttack: {
		SkillLevel1: 50, // 绿色 50%
		SkillLevel2: 30, // 蓝色 30%
		SkillLevel3: 10, // 紫色 10%
	},
	SkillTypeDefense: {
		SkillLevel1: 50, // 绿色 50%
		SkillLevel2: 30, // 蓝色 30%
		SkillLevel3: 10, // 紫色 10%
	},
	SkillTypeVampire: {
		SkillLevel1: 50, // 绿色 50%
		SkillLevel2: 30, // 蓝色 30%
		SkillLevel3: 10, // 紫色 10%
	},
}

// 随机生成种族
func GenerateRandomRace() RaceInfo {
	totalWeight := 0
	for _, weight := range RaceWeights {
		totalWeight += weight
	}

	randNum := rand.Intn(totalWeight)
	currentWeight := 0

	for race, weight := range RaceWeights {
		currentWeight += weight
		if randNum < currentWeight {
			return RaceConfigs[race]
		}
	}

	// 默认返回驼鹿
	return RaceConfigs[RaceMoose]
}

// 随机生成技能
func GenerateRandomSkill() PetSkill {
	skillTypes := []SkillType{SkillTypeAttack, SkillTypeDefense, SkillTypeVampire}
	skillType := skillTypes[rand.Intn(len(skillTypes))]

	totalWeight := 0
	for _, weight := range SkillWeights[skillType] {
		totalWeight += weight
	}

	randNum := rand.Intn(totalWeight)
	currentWeight := 0

	for level, weight := range SkillWeights[skillType] {
		currentWeight += weight
		if randNum < currentWeight {
			return SkillConfigs[skillType][level]
		}
	}

	// 默认返回1级攻击技能
	return SkillConfigs[SkillTypeAttack][SkillLevel1]
}

// NewPet 创建新宠物（不包含种族和技能，需要后续设置）
func NewPet(ownerName string) *Pet {
	personalities := []PetPersonality{
		PersonalityBrave, PersonalityGreedy, PersonalityFriendly,
		PersonalityCautious, PersonalityCurious,
	}

	petNames := []string{
		"Lucky", "Brave", "Shadow", "Spark", "Whisper",
		"Thunder", "Frost", "Blaze", "Swift", "Mystic",
	}

	baseName := petNames[len(ownerName)%len(petNames)]

	pet := &Pet{
		ID:           uuid.New().String(),
		Name:         baseName,
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

// NewPetWithRaceAndSkill 创建带有种族和技能的新宠物
func NewPetWithRaceAndSkill(ownerName string, race RaceInfo, skill PetSkill) *Pet {
	pet := NewPet(ownerName)

	// 设置种族
	pet.Race = race

	// 设置技能
	pet.Skill = skill

	// 应用种族属性加成
	pet.ApplyRaceBonuses()

	return pet
}

// ApplyRaceBonuses 应用种族属性加成
func (p *Pet) ApplyRaceBonuses() {
	p.Health += p.Race.HealthBonus
	p.MaxHealth += p.Race.HealthBonus
	p.Attack += p.Race.AttackBonus
	p.Defense += p.Race.DefenseBonus
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