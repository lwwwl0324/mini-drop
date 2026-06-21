package service

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"gorm.io/gorm"

	"mini-drop/apiserver/internal/client"
	"mini-drop/apiserver/internal/model"
	"mini-drop/apiserver/internal/storage"
)

type TaskService struct {
	dropClient *client.DropClient
	db         *gorm.DB
	storage    *storage.MinioClient
}

func NewTaskService(dropClient *client.DropClient, db *gorm.DB, storage *storage.MinioClient) *TaskService {
	return &TaskService{
		dropClient: dropClient,
		db:         db,
		storage:    storage,
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
		Status:       string(model.StatusPending),
		StatusMsg:    "任务已创建，等待下发",
		StatusReason: string(model.ReasonCreated),
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("保存任务失败: %v", err)
	}

	_, err := s.dropClient.CreateTask(ctx, targetIP, taskID, pid, duration, frequency, profilerType)
	if err != nil {
		task.Transition(model.StatusFailed, model.ReasonGRPCFailed, fmt.Sprintf("gRPC 调用失败: %v", err))
		s.db.Save(task)
		return task, err
	}

	task.Transition(model.StatusRunning, model.ReasonAgentAccepted, "任务已下发，等待 Agent 采集")
	s.db.Save(task)

	go s.handleTaskCompletion(task)

	return task, nil
}

func (s *TaskService) handleTaskCompletion(task *model.Task) {
	taskID := task.TaskID
	profilerType := task.ProfilerType
	duration := task.Duration

	waitTime := time.Duration(duration+15) * time.Second
	fmt.Printf("[Auto] 等待 %v 后检查任务 %s\n", waitTime, taskID)
	time.Sleep(waitTime)

	// 从 MinIO 检查文件是否存在
	var objectName string
	switch profilerType {
	case "1": // eBPF
		objectName = fmt.Sprintf("%s/bpftrace.log", taskID)
	case "2": // py-spy
		objectName = fmt.Sprintf("%s/pyspy.svg", taskID)
	default: // perf
		objectName = fmt.Sprintf("%s/perf.data", taskID)
	}

	fmt.Printf("[Auto] 检查 MinIO: %s\n", objectName)

	exists, err := s.storage.ObjectExists("drop-data", objectName)
	if err != nil {
		fmt.Printf("[Auto] 检查 MinIO 失败: %v\n", err)
		task.Transition(model.StatusFailed, model.ReasonPerfFailed, fmt.Sprintf("检查 MinIO 失败: %v", err))
		s.db.Save(task)
		return
	}

	if exists {
		task.Transition(model.StatusUploading, model.ReasonPerfCompleted, "采集完成，准备上传")
		s.db.Save(task)

		// 如果是 perf，生成火焰图（py-spy 和 eBPF 已经直接生成 SVG）
		if profilerType == "0" {
			cmd := exec.Command("python3", "/app/scripts/generate_flamegraph.py", taskID)
			output, cmdErr := cmd.CombinedOutput()
			if cmdErr != nil {
				task.Transition(model.StatusFailed, model.ReasonFlamegraphFailed, fmt.Sprintf("火焰图生成失败: %v", cmdErr))
				s.db.Save(task)
				fmt.Printf("[Auto] 火焰图生成失败: %v, %s\n", cmdErr, string(output))
				return
			}
		}

		flamegraphURL := fmt.Sprintf("http://localhost:9001/buckets/drop-data/browse?prefix=%s/", taskID)
		task.Transition(model.StatusDone, model.ReasonFlamegraphDone, "采集完成")
		task.FlamegraphURL = flamegraphURL
		s.db.Save(task)
		fmt.Printf("[Auto] 任务 %s 完成\n", taskID)
	} else {
		task.Transition(model.StatusFailed, model.ReasonPerfFailed, fmt.Sprintf("未在 MinIO 找到文件: %s", objectName))
		s.db.Save(task)
		fmt.Printf("[Auto] 任务 %s 失败: 未找到文件 %s\n", taskID, objectName)
	}
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
