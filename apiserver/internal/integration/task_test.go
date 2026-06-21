//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"mini-drop/apiserver/internal/client"
	"mini-drop/apiserver/internal/model"
	"mini-drop/apiserver/internal/service"
	"mini-drop/apiserver/internal/storage"
)

func setupIntegrationDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=drop password=drop123 dbname=drop port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("跳过集成测试: PostgreSQL 未运行")
	}
	return db
}

func setupStorage(t *testing.T) *storage.MinioClient {
	storage, err := storage.NewMinioClient("localhost:9000", "minioadmin", "minioadmin123", "drop-data", false)
	if err != nil {
		t.Skip("跳过集成测试: MinIO 未运行")
	}
	return storage
}

func setupDropClient(t *testing.T) *client.DropClient {
	dropClient, err := client.NewDropClient("localhost:50051")
	if err != nil {
		t.Skip("跳过集成测试: drop_server 未运行")
	}
	return dropClient
}

func TestCreateTask_Integration(t *testing.T) {
	db := setupIntegrationDB(t)
	storage := setupStorage(t)
	dropClient := setupDropClient(t)

	taskService := service.NewTaskService(dropClient, db, storage)

	ctx := context.Background()
	task, err := taskService.CreateTask(ctx, "127.0.0.1", 1, 5, 99, 0)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	if task.TaskID == "" {
		t.Error("TaskID 为空")
	}

	time.Sleep(10 * time.Second)

	updated, _ := taskService.GetTask(task.TaskID)
	if updated.Status == string(model.StatusDone) {
		t.Log("任务完成")
	}
}
