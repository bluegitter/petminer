package services

import (
	"fmt"
	"math/rand"
	"time"

	"miningpet/internal/models"
)

// ActionType 表示宠物可能的行为类型
type ActionType string

const (
	ActionExplore   ActionType = "explore"
	ActionRest      ActionType = "rest"
	ActionSocialize ActionType = "socialize"
	ActionFight     ActionType = "fight"
	ActionEat       ActionType = "eat"
	ActionIdle      ActionType = "idle"
)

// Action 表示宠物的一个行为
type Action struct {
	Type     ActionType `json:"type"`
	Priority int        `json:"priority"`
	Reason   string     `json:"reason"`
	Duration int        `json:"duration"` // 秒
}

// AIEngine AI决策引擎
type AIEngine struct {
	rand *rand.Rand
}

// NewAIEngine 创建新的AI引擎
func NewAIEngine() *AIEngine {
	return &AIEngine{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// DecideNextAction 基于宠物当前状态决定下一个行为
func (ai *AIEngine) DecideNextAction(pet *models.Pet) Action {
	actions := ai.evaluateAllActions(pet)
	
	if len(actions) == 0 {
		return Action{Type: ActionIdle, Priority: 1, Reason: "无可用行为", Duration: 30}
	}
	
	// 根据优先级和随机性选择行为
	return ai.selectAction(actions, pet)
}

// evaluateAllActions 评估所有可能的行为
func (ai *AIEngine) evaluateAllActions(pet *models.Pet) []Action {
	var actions []Action
	
	// 评估探索行为
	if pet.CanExplore() {
		priority := ai.calculateExplorePriority(pet)
		if priority > 0 {
			actions = append(actions, Action{
				Type:     ActionExplore,
				Priority: priority,
				Reason:   ai.getExploreReason(pet),
				Duration: ai.rand.Intn(60) + 30, // 30-90秒
			})
		}
	}
	
	// 评估休息行为
	if pet.CanRest() {
		priority := ai.calculateRestPriority(pet)
		if priority > 0 {
			actions = append(actions, Action{
				Type:     ActionRest,
				Priority: priority,
				Reason:   ai.getRestReason(pet),
				Duration: ai.rand.Intn(30) + 20, // 20-50秒
			})
		}
	}
	
	// 评估社交行为
	if pet.CanSocialize() {
		priority := ai.calculateSocializePriority(pet)
		if priority > 0 {
			actions = append(actions, Action{
				Type:     ActionSocialize,
				Priority: priority,
				Reason:   ai.getSocializeReason(pet),
				Duration: ai.rand.Intn(40) + 25, // 25-65秒
			})
		}
	}
	
	// 评估进食行为
	priority := ai.calculateEatPriority(pet)
	if priority > 0 {
		actions = append(actions, Action{
			Type:     ActionEat,
			Priority: priority,
			Reason:   ai.getEatReason(pet),
			Duration: ai.rand.Intn(20) + 10, // 10-30秒
		})
	}
	
	return actions
}

// 计算探索行为优先级
func (ai *AIEngine) calculateExplorePriority(pet *models.Pet) int {
	priority := 50 // 基础优先级
	
	// 基于性格调整
	switch pet.Personality {
	case models.PersonalityCurious:
		priority += 30
	case models.PersonalityBrave:
		priority += 20
	case models.PersonalityGreedy:
		priority += 15
	case models.PersonalityCautious:
		priority -= 10
	}
	
	// 基于体力调整
	energyPercent := float64(pet.Energy) / float64(pet.MaxEnergy)
	if energyPercent < 0.3 {
		priority -= 40
	} else if energyPercent > 0.8 {
		priority += 20
	}
	
	// 基于心情调整
	priority += int(pet.GetMoodInfluence() * 10)
	
	// 基于饱食度调整
	if pet.Hunger < 40 {
		priority -= 30
	}
	
	if priority < 0 {
		priority = 0
	}
	
	return priority
}

// 计算休息行为优先级
func (ai *AIEngine) calculateRestPriority(pet *models.Pet) int {
	priority := 20 // 基础优先级
	
	// 基于体力调整
	energyPercent := float64(pet.Energy) / float64(pet.MaxEnergy)
	if energyPercent < 0.3 {
		priority += 80
	} else if energyPercent < 0.5 {
		priority += 40
	}
	
	// 基于健康状况调整
	healthPercent := float64(pet.Health) / float64(pet.MaxHealth)
	if healthPercent < 0.5 {
		priority += 60
	}
	
	// 基于心情调整
	if pet.Mood == models.MoodTired {
		priority += 50
	}
	
	// 基于性格调整
	if pet.Personality == models.PersonalityCautious {
		priority += 20
	}
	
	return priority
}

// 计算社交行为优先级
func (ai *AIEngine) calculateSocializePriority(pet *models.Pet) int {
	priority := 30 // 基础优先级
	
	// 基于社交度调整
	if pet.Social < 30 {
		priority += 60
	} else if pet.Social < 50 {
		priority += 30
	}
	
	// 基于性格调整
	switch pet.Personality {
	case models.PersonalityFriendly:
		priority += 40
	case models.PersonalityCurious:
		priority += 20
	case models.PersonalityCautious:
		priority -= 10
	}
	
	// 基于心情调整
	if pet.Mood == models.MoodSad {
		priority += 30
	}
	
	return priority
}

// 计算进食行为优先级
func (ai *AIEngine) calculateEatPriority(pet *models.Pet) int {
	// 如果饱食度已达到90以上，不需要进食
	if pet.Hunger >= 90 {
		return 0
	}
	
	priority := 0
	
	// 基于饱食度调整
	if pet.Hunger < 20 {
		priority = 100 // 极高优先级
	} else if pet.Hunger < 40 {
		priority = 70
	} else if pet.Hunger < 60 {
		priority = 30
	} else if pet.Hunger < 80 {
		priority = 10 // 轻微饥饿
	}
	
	// 基于性格调整（但不能导致已饱食时还要进食）
	if pet.Personality == models.PersonalityGreedy && pet.Hunger < 85 {
		priority += 15
	}
	
	return priority
}

// selectAction 基于优先级和随机性选择行为
func (ai *AIEngine) selectAction(actions []Action, pet *models.Pet) Action {
	if len(actions) == 1 {
		return actions[0]
	}
	
	// 计算加权随机选择
	totalWeight := 0
	for _, action := range actions {
		totalWeight += action.Priority
	}
	
	if totalWeight == 0 {
		return actions[ai.rand.Intn(len(actions))]
	}
	
	randomValue := ai.rand.Intn(totalWeight)
	currentWeight := 0
	
	for _, action := range actions {
		currentWeight += action.Priority
		if randomValue < currentWeight {
			return action
		}
	}
	
	return actions[len(actions)-1]
}

// 获取各种行为的原因描述
func (ai *AIEngine) getExploreReason(pet *models.Pet) string {
	reasons := []string{
		fmt.Sprintf("%s 想要寻找新的冒险", pet.Name),
		fmt.Sprintf("%s 对未知的地方充满好奇", pet.Name),
		fmt.Sprintf("%s 希望找到一些宝藏", pet.Name),
	}
	
	switch pet.Personality {
	case models.PersonalityCurious:
		reasons = append(reasons, fmt.Sprintf("%s 的好奇心驱使着探索", pet.Name))
	case models.PersonalityBrave:
		reasons = append(reasons, fmt.Sprintf("%s 勇敢地踏上冒险之路", pet.Name))
	case models.PersonalityGreedy:
		reasons = append(reasons, fmt.Sprintf("%s 想要寻找更多财富", pet.Name))
	}
	
	return reasons[ai.rand.Intn(len(reasons))]
}

func (ai *AIEngine) getRestReason(pet *models.Pet) string {
	if pet.Energy < 30 {
		return fmt.Sprintf("%s 感到很疲惫，需要休息", pet.Name)
	}
	if pet.Health < 50 {
		return fmt.Sprintf("%s 需要恢复体力", pet.Name)
	}
	return fmt.Sprintf("%s 想要放松一下", pet.Name)
}

func (ai *AIEngine) getSocializeReason(pet *models.Pet) string {
	if pet.Social < 30 {
		return fmt.Sprintf("%s 感到孤独，想要交朋友", pet.Name)
	}
	return fmt.Sprintf("%s 想要和其他宠物互动", pet.Name)
}

func (ai *AIEngine) getEatReason(pet *models.Pet) string {
	if pet.Hunger < 30 {
		return fmt.Sprintf("%s 感到很饿，急需进食", pet.Name)
	}
	return fmt.Sprintf("%s 想要补充体力", pet.Name)
}