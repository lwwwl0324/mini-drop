package service

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"gorm.io/gorm"

	"mini-drop/apiserver/internal/client"
	"mini-drop/apiserver/internal/model"
)

type TaskService struct {
	dropClient *client.DropClient
	db         *gorm.DB
}

func NewTaskService(dropClient *client.DropClient, db *gorm.DB) *TaskService {
	return &TaskService{
		dropClient: dropClient,
		db:         db,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, targetIP string, pid, duration, frequency int, profilerType int) (*model.Task, error) {
	taskID := fmt.Sprintf("task_%d", time.Now().Unix())

	task := &model.Task{
		TaskID:       taskID,
		TargetIP:     targetIP,
		PID:          pid,
		Duration:     duration,
		Frequency:    frequency,
		ProfilerType: fmt.Sprintf("%d", profilerType),
		Status:       "pending",
		StatusMsg:    "任务已创建，等待下发",
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("保存任务失败: %v", err)
	}

	_, err := s.dropClient.CreateTask(ctx, targetIP, taskID, pid, duration, frequency, profilerType)
	if err != nil {
		task.Status = "failed"
		task.StatusMsg = err.Error()
		s.db.Save(task)
		return task, err
	}

	task.Status = "running"
	task.StatusMsg = "任务已下发，Agent 正在采集"
	s.db.Save(task)

	go s.autoGenerateFlamegraph(taskID, duration, pid, profilerType)

	return task, nil
}

func (s *TaskService) autoGenerateFlamegraph(taskID string, duration, pid, profilerType int) {
	waitTime := time.Duration(duration+15) * time.Second
	fmt.Printf("[Auto] 等待 %v 后生成火焰图\n", waitTime)
	time.Sleep(waitTime)

	var cmd *exec.Cmd
	if profilerType == 1 {
		// eBPF 采集：直接显示日志
		logFile := fmt.Sprintf("/tmp/bpftrace_%s.log", taskID)
		cmd = exec.Command("bash", "-c", fmt.Sprintf("cat %s", logFile))
	} else {
		cmd = exec.Command("python3", "/home/lwl/mini-drop/scripts/generate_flamegraph.py", taskID)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[Auto] 火焰图生成失败: %v, %s\n", err, string(output))
		return
	}

	fmt.Printf("[Auto] 火焰图生成成功: %s\n", taskID)

	flamegraphURL := fmt.Sprintf("http://localhost:9001/buckets/drop-data/browse?prefix=%s/", taskID)
	s.db.Model(&model.Task{}).Where("task_id = ?", taskID).Updates(map[string]interface{}{
		"status":         "done",
		"status_msg":     "采集完成，火焰图已生成",
		"flamegraph_url": flamegraphURL,
	})
}

func (s *TaskService) GetTask(taskID string) (*model.Task, error) {
	var task model.Task
	err := s.db.Where("task_id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *TaskService) ListTasks() ([]model.Task, error) {
	var tasks []model.Task
	err := s.db.Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}
