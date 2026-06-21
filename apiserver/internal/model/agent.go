package model

import (
	"time"
	"gorm.io/gorm"
)

// Agent Agent 信息
type Agent struct {
	ID            uint           `gorm:"primaryKey"`
	AgentID       string         `gorm:"uniqueIndex;size:64;not null" json:"agent_id"`
	Hostname      string         `gorm:"size:128" json:"hostname"`
	IPAddr        string         `gorm:"size:64;not null" json:"ip_addr"`
	Version       string         `gorm:"size:32" json:"version"`
	Status        string         `gorm:"size:32;default:'online'" json:"status"` // online / offline
	LastHeartbeat time.Time      `json:"last_heartbeat"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Agent) TableName() string {
	return "agents"
}
