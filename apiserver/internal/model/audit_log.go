package model

import (
	"time"
	"gorm.io/gorm"
)

// AuditLog 审计日志
type AuditLog struct {
	ID          uint           `gorm:"primaryKey"`
	AgentID     string         `gorm:"size:64;not null" json:"agent_id"`
	EventType   string         `gorm:"size:32;not null" json:"event_type"` // online / offline / heartbeat / task_accepted / task_completed
	EventDetail string         `gorm:"size:512" json:"event_detail"`
	IPAddr      string         `gorm:"size:64" json:"ip_addr"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
