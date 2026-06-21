//go:build integration
// +build integration

package service

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"mini-drop/apiserver/internal/model"
)

// 这个测试需要真实的 drop_server 和 MinIO 环境
// 使用 build tag "integration" 隔离

func setupIntegrationDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("无法创建测试数据库: %v", err)
	}
	db.AutoMigrate(&model.Task{})
	return db
}

func TestCreateTask_Integration(t *testing.T) {
	// 跳过集成测试（需要真实环境）
	t.Skip("跳过集成测试，需要真实 drop_server 和 MinIO")
}
