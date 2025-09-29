package database

import (
	"time"
	"encoding/json"
)

// DBPet 数据库宠物模型
type DBPet struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	Name         string    `gorm:"size:50;not null" json:"name"`
	Owner        string    `gorm:"size:50;not null;index" json:"owner"`
	Personality  string    `gorm:"size:20;not null" json:"personality"`
	Level        int       `gorm:"default:1" json:"level"`
	Experience   int       `gorm:"default:0" json:"experience"`
	Health       int       `gorm:"default:100" json:"health"`
	MaxHealth    int       `gorm:"default:100" json:"max_health"`
	Energy       int       `gorm:"default:100" json:"energy"`
	MaxEnergy    int       `gorm:"default:100" json:"max_energy"`
	Hunger       int       `gorm:"default:80" json:"hunger"`
	Social       int       `gorm:"default:50" json:"social"`
	Mood         string    `gorm:"size:20;default:'普通'" json:"mood"`
	Attack       int       `gorm:"default:10" json:"attack"`
	Defense      int       `gorm:"default:5" json:"defense"`
	Coins        int       `gorm:"default:0" json:"coins"`
	Location     string    `gorm:"size:100;default:'起始村庄'" json:"location"`
	Status       string    `gorm:"size:20;default:'等待中'" json:"status"`
	Memory       string    `gorm:"type:text" json:"memory"`       // JSON存储
	Friends      string    `gorm:"type:text" json:"friends"`      // JSON存储
	LastActivity time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"last_activity"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// DBEvent 数据库事件模型
type DBEvent struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	PetID     string    `gorm:"size:36;not null;index" json:"pet_id"`
	PetName   string    `gorm:"size:50;not null" json:"pet_name"`
	Type      string    `gorm:"size:20;not null;index" json:"type"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	Data      string    `gorm:"type:text" json:"data"`      // JSON存储EventData
	Timestamp time.Time `gorm:"not null;index" json:"timestamp"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName 指定表名
func (DBPet) TableName() string {
	return "pets"
}

func (DBEvent) TableName() string {
	return "events"
}

// 辅助方法：JSON序列化/反序列化
func (p *DBPet) SetMemory(memory []string) error {
	if memory == nil {
		p.Memory = "[]"
		return nil
	}
	data, err := json.Marshal(memory)
	if err != nil {
		return err
	}
	p.Memory = string(data)
	return nil
}

func (p *DBPet) GetMemory() ([]string, error) {
	if p.Memory == "" {
		return []string{}, nil
	}
	var memory []string
	err := json.Unmarshal([]byte(p.Memory), &memory)
	return memory, err
}

func (p *DBPet) SetFriends(friends []string) error {
	if friends == nil {
		p.Friends = "[]"
		return nil
	}
	data, err := json.Marshal(friends)
	if err != nil {
		return err
	}
	p.Friends = string(data)
	return nil
}

func (p *DBPet) GetFriends() ([]string, error) {
	if p.Friends == "" {
		return []string{}, nil
	}
	var friends []string
	err := json.Unmarshal([]byte(p.Friends), &friends)
	return friends, err
}

func (e *DBEvent) SetEventData(data interface{}) error {
	if data == nil {
		e.Data = "{}"
		return nil
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	e.Data = string(jsonData)
	return nil
}

func (e *DBEvent) GetEventData() (map[string]interface{}, error) {
	if e.Data == "" {
		return make(map[string]interface{}), nil
	}
	var data map[string]interface{}
	err := json.Unmarshal([]byte(e.Data), &data)
	return data, err
}