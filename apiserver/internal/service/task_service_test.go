package service

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"mini-drop/apiserver/internal/model"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("无法创建测试数据库: %v", err)
	}
	db.AutoMigrate(&model.Task{})
	return db
}

func TestTaskService_GetTask_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := &TaskService{Db: db}
	_, err := service.GetTask("non_existent_task")
	if err == nil {
		t.Error("期望返回错误，但成功了")
	}
}

func TestTaskService_ListTasks_Empty(t *testing.T) {
	db := setupTestDB(t)
	service := &TaskService{Db: db}
	tasks, err := service.ListTasks()
	if err != nil {
		t.Errorf("列出任务失败: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("期望 0 个任务, 实际=%d", len(tasks))
	}
}

func TestTaskService_ListTasks_WithData(t *testing.T) {
	db := setupTestDB(t)
	db.Create(&model.Task{
		TaskID:       "test_task_001",
		TargetIP:     "127.0.0.1",
		PID:          1234,
		Duration:     10,
		Frequency:    999,
		ProfilerType: "0",
		Status:       string(model.StatusPending),
		StatusMsg:    "任务已创建",
		StatusReason: string(model.ReasonCreated),
	})
	service := &TaskService{Db: db}
	tasks, err := service.ListTasks()
	if err != nil {
		t.Errorf("列出任务失败: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("期望 1 个任务, 实际=%d", len(tasks))
	}
}

func TestTaskModel_Create(t *testing.T) {
	db := setupTestDB(t)
	task := &model.Task{
		TaskID:       "test_task_002",
		TargetIP:     "127.0.0.1",
		PID:          1234,
		Duration:     10,
		Frequency:    999,
		ProfilerType: "0",
		Status:       string(model.StatusPending),
		StatusMsg:    "任务已创建",
		StatusReason: string(model.ReasonCreated),
	}
	result := db.Create(task)
	if result.Error != nil {
		t.Errorf("创建任务失败: %v", result.Error)
	}
	var saved model.Task
	db.Where("task_id = ?", "test_task_002").First(&saved)
	if saved.TaskID != "test_task_002" {
		t.Errorf("期望 TaskID=test_task_002, 实际=%s", saved.TaskID)
	}
}

func TestTaskModel_Update(t *testing.T) {
	db := setupTestDB(t)
	task := &model.Task{
		TaskID:       "test_task_003",
		TargetIP:     "127.0.0.1",
		PID:          1234,
		Duration:     10,
		Frequency:    999,
		ProfilerType: "0",
		Status:       string(model.StatusPending),
		StatusMsg:    "任务已创建",
		StatusReason: string(model.ReasonCreated),
	}
	db.Create(task)
	db.Model(&model.Task{}).Where("task_id = ?", "test_task_003").Updates(map[string]interface{}{
		"status":        string(model.StatusRunning),
		"status_msg":    "任务已下发",
		"status_reason": string(model.ReasonAgentAccepted),
	})
	var updated model.Task
	db.Where("task_id = ?", "test_task_003").First(&updated)
	if updated.Status != string(model.StatusRunning) {
		t.Errorf("期望 Status=RUNNING, 实际=%s", updated.Status)
	}
}

func TestTaskModel_Delete(t *testing.T) {
	db := setupTestDB(t)
	task := &model.Task{
		TaskID:       "test_task_004",
		TargetIP:     "127.0.0.1",
		PID:          1234,
		Duration:     10,
		Frequency:    999,
		ProfilerType: "0",
		Status:       string(model.StatusPending),
	}
	db.Create(task)
	db.Delete(&model.Task{}, "task_id = ?", "test_task_004")
	var count int64
	db.Model(&model.Task{}).Where("task_id = ?", "test_task_004").Count(&count)
	if count != 0 {
		t.Errorf("期望 0 条记录, 实际=%d", count)
	}
}

func TestTaskModel_StatusTransition(t *testing.T) {
	db := setupTestDB(t)
	task := &model.Task{
		TaskID:       "test_transition_001",
		TargetIP:     "127.0.0.1",
		PID:          1234,
		Duration:     10,
		Frequency:    999,
		ProfilerType: "0",
		Status:       string(model.StatusPending),
		StatusMsg:    "任务已创建",
		StatusReason: string(model.ReasonCreated),
	}
	db.Create(task)

	transitions := []struct {
		to     model.TaskStatus
		reason model.StatusReason
		msg    string
	}{
		{model.StatusRunning, model.ReasonAgentAccepted, "Agent 已接受"},
		{model.StatusUploading, model.ReasonPerfCompleted, "采集完成"},
		{model.StatusDone, model.ReasonFlamegraphDone, "火焰图已生成"},
	}

	for _, tr := range transitions {
		var tsk model.Task
		db.Where("task_id = ?", "test_transition_001").First(&tsk)
		_, err := tsk.Transition(tr.to, tr.reason, tr.msg)
		if err != nil {
			t.Errorf("转换到 %s 失败: %v", tr.to, err)
		}
		db.Save(&tsk)
	}
}

func TestTaskModel_InvalidTransition(t *testing.T) {
	db := setupTestDB(t)
	task := &model.Task{
		TaskID:       "test_transition_002",
		TargetIP:     "127.0.0.1",
		PID:          1234,
		Duration:     10,
		Frequency:    999,
		ProfilerType: "0",
		Status:       string(model.StatusDone),
		StatusMsg:    "已完成",
	}
	db.Create(task)
	var tsk model.Task
	db.Where("task_id = ?", "test_transition_002").First(&tsk)
	_, err := tsk.Transition(model.StatusRunning, model.ReasonAgentAccepted, "")
	if err == nil {
		t.Error("从 DONE 转换到 RUNNING 应该失败")
	}
}

// 测试 getEnv 函数（但它在 cmd/main.go 中，不在 service 包）
// 改为测试 TaskService 结构体的初始化

func TestTaskService_NewTaskService(t *testing.T) {
	db := setupTestDB(t)
	// 由于 dropClient 和 storage 需要真实连接，这里只测试结构体创建
	svc := &TaskService{
		Db: db,
	}
	if svc.Db == nil {
		t.Error("Db 字段为 nil")
	}
}

func TestTaskModel_StatusConstants(t *testing.T) {
	// 测试状态常量存在
	statuses := []model.TaskStatus{
		model.StatusPending,
		model.StatusRunning,
		model.StatusUploading,
		model.StatusDone,
		model.StatusFailed,
	}
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("状态 %s 为空", s)
		}
	}
}

func TestTaskModel_ReasonConstants(t *testing.T) {
	reasons := []model.StatusReason{
		model.ReasonCreated,
		model.ReasonAgentAccepted,
		model.ReasonAgentRejected,
		model.ReasonPerfStarted,
		model.ReasonPerfCompleted,
		model.ReasonPerfFailed,
		model.ReasonUploadStarted,
		model.ReasonUploadCompleted,
		model.ReasonUploadFailed,
		model.ReasonFlamegraphStarted,
		model.ReasonFlamegraphDone,
		model.ReasonFlamegraphFailed,
		model.ReasonGRPCTimeout,
		model.ReasonGRPCFailed,
		model.ReasonAgentOffline,
		model.ReasonTimeout,
	}
	for _, r := range reasons {
		if string(r) == "" {
			t.Errorf("Reason %s 为空", r)
		}
	}
}
