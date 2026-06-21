package model

import (
	"testing"
	"time"
)

func TestTaskTransition(t *testing.T) {
	task := &Task{
		TaskID:       "test_task_001",
		TargetIP:     "127.0.0.1",
		PID:          1234,
		Duration:     10,
		Frequency:    999,
		ProfilerType: "0",
		Status:       string(StatusPending),
		StatusMsg:    "任务已创建",
		StatusReason: string(ReasonCreated),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 测试: PENDING -> RUNNING
	transition, err := task.Transition(StatusRunning, ReasonAgentAccepted, "Agent 已接受")
	if err != nil {
		t.Errorf("转换失败: %v", err)
	}
	if transition.From != StatusPending {
		t.Errorf("期望 From=PENDING, 实际=%s", transition.From)
	}
	if transition.To != StatusRunning {
		t.Errorf("期望 To=RUNNING, 实际=%s", transition.To)
	}
	if task.Status != string(StatusRunning) {
		t.Errorf("期望 Status=RUNNING, 实际=%s", task.Status)
	}
	if task.StatusReason != string(ReasonAgentAccepted) {
		t.Errorf("期望 StatusReason=agent_accepted, 实际=%s", task.StatusReason)
	}

	// 测试: RUNNING -> UPLOADING
	transition, err = task.Transition(StatusUploading, ReasonPerfCompleted, "perf 采集完成")
	if err != nil {
		t.Errorf("转换失败: %v", err)
	}
	if transition.From != StatusRunning {
		t.Errorf("期望 From=RUNNING, 实际=%s", transition.From)
	}
	if transition.To != StatusUploading {
		t.Errorf("期望 To=UPLOADING, 实际=%s", transition.To)
	}

	// 测试: UPLOADING -> DONE
	transition, err = task.Transition(StatusDone, ReasonFlamegraphDone, "火焰图已生成")
	if err != nil {
		t.Errorf("转换失败: %v", err)
	}
	if transition.From != StatusUploading {
		t.Errorf("期望 From=UPLOADING, 实际=%s", transition.From)
	}
	if transition.To != StatusDone {
		t.Errorf("期望 To=DONE, 实际=%s", transition.To)
	}
}

func TestInvalidTransition(t *testing.T) {
	task := &Task{
		TaskID: "test_task_002",
		Status: string(StatusDone),
	}

	// 尝试从终态 DONE 转换 -> 应该失败
	_, err := task.Transition(StatusRunning, ReasonAgentAccepted, "")
	if err == nil {
		t.Error("期望从 DONE 转换失败，但成功了")
	}
}
