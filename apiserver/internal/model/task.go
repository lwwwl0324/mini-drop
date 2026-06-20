package model

import (
	"time"
	"gorm.io/gorm"
)

// Task 任务表
type Task struct {
	ID           uint           `gorm:"primaryKey"`
	TaskID       string         `gorm:"uniqueIndex;size:64;not null" json:"task_id"`
	TargetIP     string         `gorm:"size:64;not null" json:"target_ip"`
	PID          int            `json:"pid"`
	Duration     int            `json:"duration"`
	Frequency    int            `json:"frequency"`
	ProfilerType string         `gorm:"size:32;default:'perf'" json:"profiler_type"`
	Status       string         `gorm:"size:32;default:'pending'" json:"status"`
	StatusMsg    string         `gorm:"size:256" json:"status_msg"`
	FlamegraphURL string        `gorm:"size:512" json:"flamegraph_url"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}
