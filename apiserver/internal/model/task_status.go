package model

import (
	"fmt"
	"time"
)

// TaskStatus 任务状态枚举
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusUploading TaskStatus = "uploading"
	StatusDone      TaskStatus = "done"
	StatusFailed    TaskStatus = "failed"
)

// StatusReason 状态变化原因
type StatusReason string

const (
	ReasonCreated           StatusReason = "task_created"
	ReasonAgentAccepted     StatusReason = "agent_accepted"
	ReasonAgentRejected     StatusReason = "agent_rejected"
	ReasonPerfStarted       StatusReason = "perf_started"
	ReasonPerfCompleted     StatusReason = "perf_completed"
	ReasonPerfFailed        StatusReason = "perf_failed"
	ReasonEBPFStarted       StatusReason = "ebpf_started"
	ReasonEBPFCompleted     StatusReason = "ebpf_completed"
	ReasonEBPFFailed        StatusReason = "ebpf_failed"
	ReasonUploadStarted     StatusReason = "upload_started"
	ReasonUploadCompleted   StatusReason = "upload_completed"
	ReasonUploadFailed      StatusReason = "upload_failed"
	ReasonFlamegraphStarted StatusReason = "flamegraph_started"
	ReasonFlamegraphDone    StatusReason = "flamegraph_generated"
	ReasonFlamegraphFailed  StatusReason = "flamegraph_failed"
	ReasonGRPCTimeout       StatusReason = "grpc_timeout"
	ReasonGRPCFailed        StatusReason = "grpc_failed"
	ReasonAgentOffline      StatusReason = "agent_offline"
	ReasonTimeout           StatusReason = "task_timeout"
)

// StatusTransition 状态迁移记录
type StatusTransition struct {
	From   TaskStatus   `json:"from"`
	To     TaskStatus   `json:"to"`
	Reason StatusReason `json:"reason"`
	Msg    string       `json:"msg"`
	Time   time.Time    `json:"time"`
}

// Transition 执行状态迁移
func (t *Task) Transition(to TaskStatus, reason StatusReason, msg string) (*StatusTransition, error) {
	from := TaskStatus(t.Status)
	
	// 验证状态迁移是否合法
	if !isValidTransition(from, to) {
		return nil, fmt.Errorf("非法状态迁移: %s -> %s", from, to)
	}
	
	// 执行迁移
	t.Status = string(to)
	t.StatusReason = string(reason)
	if msg != "" {
		t.StatusMsg = msg
	}
	t.UpdatedAt = time.Now()
	
	return &StatusTransition{
		From:   from,
		To:     to,
		Reason: reason,
		Msg:    msg,
		Time:   t.UpdatedAt,
	}, nil
}

// isValidTransition 验证状态迁移是否合法
func isValidTransition(from, to TaskStatus) bool {
	// 定义合法的状态迁移
	validTransitions := map[TaskStatus][]TaskStatus{
		StatusPending:   {StatusRunning, StatusFailed},
		StatusRunning:   {StatusUploading, StatusFailed},
		StatusUploading: {StatusDone, StatusFailed},
		StatusDone:      {}, // 终态，不能迁移
		StatusFailed:    {}, // 终态，不能迁移
	}
	
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	
	for _, allowedTo := range allowed {
		if allowedTo == to {
			return true
		}
	}
	return false
}
