package models

import (
	"time"
)

type EventType string

const (
	EventExplore     EventType = "explore"
	EventBattle      EventType = "battle"
	EventDiscovery   EventType = "discovery"
	EventSocial      EventType = "social"
	EventReward      EventType = "reward"
	EventLevelUp     EventType = "level_up"
	EventRareFind    EventType = "rare_find"
)

type Event struct {
	ID        string    `json:"id"`
	PetID     string    `json:"pet_id"`
	PetName   string    `json:"pet_name"`
	Type      EventType `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Data      EventData `json:"data"`
}

type EventData struct {
	Location     string `json:"location,omitempty"`
	Experience   int    `json:"experience,omitempty"`
	Coins        int    `json:"coins,omitempty"`
	Items        []Item `json:"items,omitempty"`
	Enemy        string `json:"enemy,omitempty"`
	Damage       int    `json:"damage,omitempty"`
	IsVictory    bool   `json:"is_victory,omitempty"`
	FriendName   string `json:"friend_name,omitempty"`
	NewLevel     int    `json:"new_level,omitempty"`
	RareItem     string `json:"rare_item,omitempty"`
}

type Monster struct {
	Name     string `json:"name"`
	Health   int    `json:"health"`
	Attack   int    `json:"attack"`
	Defense  int    `json:"defense"`
	ExpReward int   `json:"exp_reward"`
	CoinReward int  `json:"coin_reward"`
}

var Monsters = []Monster{
	{Name: "野猪", Health: 30, Attack: 8, Defense: 2, ExpReward: 15, CoinReward: 5},
	{Name: "森林狼", Health: 40, Attack: 12, Defense: 3, ExpReward: 20, CoinReward: 8},
	{Name: "山贼", Health: 50, Attack: 15, Defense: 5, ExpReward: 30, CoinReward: 15},
	{Name: "巨型蜘蛛", Health: 60, Attack: 18, Defense: 4, ExpReward: 35, CoinReward: 12},
	{Name: "洞穴熊", Health: 80, Attack: 20, Defense: 8, ExpReward: 50, CoinReward: 25},
}

var Locations = []string{
	"北方森林", "东部山脉", "南方沼泽", "西部草原", "神秘洞穴",
	"古老废墟", "水晶矿洞", "魔法森林", "暗影峡谷", "天空之城遗址",
}