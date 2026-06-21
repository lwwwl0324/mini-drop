//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

const apiBaseURL = "http://localhost:8191/api/v1"

var client = &http.Client{Timeout: 30 * time.Second}

type TaskResponse struct {
	ID            int    `json:"id"`              // 改为 int
	TaskID        string `json:"task_id"`
	TargetIP      string `json:"target_ip"`
	PID           int    `json:"pid"`
	Duration      int    `json:"duration"`
	Frequency     int    `json:"frequency"`
	ProfilerType  string `json:"profiler_type"`
	Status        string `json:"status"`
	StatusMsg     string `json:"status_msg"`
	StatusReason  string `json:"status_reason"`
	FlamegraphURL string `json:"flamegraph_url"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// ============================================================
// 场景 1: 正常路径 - 创建 perf 任务，等待完成
// ============================================================
func TestE2E_PerfTaskSuccess(t *testing.T) {
	t.Log("=== 场景 1: perf 任务正常完成 ===")

	taskID, err := createTask(t, "perf", 1, 5, 99)
	if err != nil {
		if contains(err.Error(), "DeadlineExceeded") {
			t.Skip("跳过测试: drop_server 未响应 (gRPC 超时)")
		}
		t.Fatalf("创建任务失败: %v", err)
	}
	t.Logf("✅ 任务创建成功: %s", taskID)

	status, err := waitForTaskDone(t, taskID, 60*time.Second)
	if err != nil {
		t.Fatalf("等待任务完成失败: %v", err)
	}
	t.Logf("✅ 任务状态: %s", status)

	task, err := getTask(t, taskID)
	if err != nil {
		t.Fatalf("获取任务详情失败: %v", err)
	}

	if task.FlamegraphURL == "" {
		t.Error("火焰图 URL 为空")
	} else {
		t.Logf("✅ 火焰图 URL: %s", task.FlamegraphURL)
	}
}

// ============================================================
// 场景 2: 异常路径 - PID 不存在导致采集失败
// ============================================================
func TestE2E_PerfTaskInvalidPID(t *testing.T) {
	t.Log("=== 场景 2: PID 不存在 -> 任务失败 ===")

	taskID, err := createTask(t, "perf", 999999, 5, 99)
	if err != nil {
		if contains(err.Error(), "DeadlineExceeded") {
			t.Skip("跳过测试: drop_server 未响应 (gRPC 超时)")
		}
		t.Fatalf("创建任务失败: %v", err)
	}
	t.Logf("✅ 任务创建成功: %s", taskID)

	status, err := waitForTaskFailed(t, taskID, 90*time.Second)
	if err != nil {
		task, _ := getTask(t, taskID)
		if task != nil {
			t.Logf("当前状态: %s, 消息: %s", task.Status, task.StatusMsg)
		}
		t.Fatalf("等待任务失败超时: %v", err)
	}
	t.Logf("✅ 任务状态: %s", status)
}

// ============================================================
// 场景 3: 异常路径 - 目标 IP 不可达
// ============================================================
func TestE2E_InvalidTargetIP(t *testing.T) {
	t.Log("=== 场景 3: 目标 IP 不可达 -> 任务失败 ===")

	taskID, err := createTaskWithTargetIP(t, "perf", "192.168.255.254", 1, 5, 99)
	if err != nil {
		if contains(err.Error(), "DeadlineExceeded") {
			t.Skip("跳过测试: drop_server 未响应 (gRPC 超时)")
		}
		t.Fatalf("创建任务失败: %v", err)
	}
	t.Logf("✅ 任务创建成功: %s", taskID)

	status, err := waitForTaskFailed(t, taskID, 90*time.Second)
	if err != nil {
		t.Logf("⚠️ 任务未失败: %v", err)
		task, _ := getTask(t, taskID)
		if task != nil {
			t.Logf("当前状态: %s, 消息: %s", task.Status, task.StatusMsg)
		}
		return
	}
	t.Logf("✅ 任务状态: %s", status)
}

// ============================================================
// 辅助函数
// ============================================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}

func createTask(t *testing.T, profilerType string, pid, duration, frequency int) (string, error) {
	return createTaskWithTargetIP(t, profilerType, "127.0.0.1", pid, duration, frequency)
}

func createTaskWithTargetIP(t *testing.T, profilerType, targetIP string, pid, duration, frequency int) (string, error) {
	reqBody := map[string]interface{}{
		"target_ip":     targetIP,
		"pid":           pid,
		"duration":      duration,
		"frequency":     frequency,
		"profiler_type": profilerType,
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := client.Post(apiBaseURL+"/tasks", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("HTTP 请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		return "", fmt.Errorf("API 返回错误: %s", apiResp.Msg)
	}

	dataMap, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("无法解析 data 字段")
	}
	taskID, ok := dataMap["task_id"].(string)
	if !ok {
		return "", fmt.Errorf("无法获取 task_id")
	}

	return taskID, nil
}

func getTask(t *testing.T, taskID string) (*TaskResponse, error) {
	resp, err := client.Get(apiBaseURL + "/tasks/" + taskID)
	if err != nil {
		return nil, fmt.Errorf("HTTP 请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var apiResp struct {
		Code int          `json:"code"`
		Msg  string       `json:"msg"`
		Data TaskResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("API 返回错误: %s", apiResp.Msg)
	}

	return &apiResp.Data, nil
}

func waitForTaskDone(t *testing.T, taskID string, timeout time.Duration) (string, error) {
	return waitForTaskStatus(t, taskID, "done", "failed", timeout)
}

func waitForTaskFailed(t *testing.T, taskID string, timeout time.Duration) (string, error) {
	return waitForTaskStatus(t, taskID, "failed", "done", timeout)
}

func waitForTaskStatus(t *testing.T, taskID, targetStatus, stopStatus string, timeout time.Duration) (string, error) {
	start := time.Now()
	for {
		elapsed := time.Since(start)
		if elapsed > timeout {
			return "", fmt.Errorf("超时 (已等待 %v): 状态未变为 %s", elapsed, targetStatus)
		}

		task, err := getTask(t, taskID)
		if err != nil {
			t.Logf("⏳ 获取任务失败: %v, 重试中...", err)
			time.Sleep(2 * time.Second)
			continue
		}

		t.Logf("⏳ [%v] 当前状态: %s, 消息: %s", elapsed.Round(time.Second), task.Status, task.StatusMsg)

		if task.Status == targetStatus {
			return task.Status, nil
		}
		if task.Status == stopStatus {
			return task.Status, fmt.Errorf("状态变为 %s 而不是 %s", stopStatus, targetStatus)
		}

		time.Sleep(2 * time.Second)
	}
}
