//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	apiBaseURL = "http://localhost:8191/api/v1"
)

// HTTP 客户端
var client = &http.Client{Timeout: 30 * time.Second}

// ============================================================
// 场景 1: 正常路径 - 创建 perf 任务，等待完成
// ============================================================
func TestE2E_PerfTaskSuccess(t *testing.T) {
	t.Log("=== 场景 1: perf 任务正常完成 ===")

	// 1. 创建任务
	taskID, err := createTask(t, "perf", 1, 5, 99)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}
	t.Logf("✅ 任务创建成功: %s", taskID)

	// 2. 等待任务完成（最多 60 秒）
	status, err := waitForTaskDone(t, taskID, 60*time.Second)
	if err != nil {
		t.Fatalf("等待任务完成失败: %v", err)
	}
	t.Logf("✅ 任务状态: %s", status)

	// 3. 验证火焰图 URL 存在
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

	// 1. 创建任务（PID 不存在）
	taskID, err := createTask(t, "perf", 999999, 5, 99)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}
	t.Logf("✅ 任务创建成功: %s", taskID)

	// 2. 等待任务失败（最多 60 秒）
	status, err := waitForTaskFailed(t, taskID, 60*time.Second)
	if err != nil {
		t.Fatalf("等待任务失败超时: %v", err)
	}
	t.Logf("✅ 任务状态: %s", status)

	// 3. 验证状态信息包含错误描述
	task, err := getTask(t, taskID)
	if err != nil {
		t.Fatalf("获取任务详情失败: %v", err)
	}

	if task.Status != "failed" {
		t.Errorf("期望状态 failed, 实际=%s", task.Status)
	}
	if task.StatusMsg == "" {
		t.Error("状态信息为空")
	} else {
		t.Logf("✅ 状态信息: %s", task.StatusMsg)
	}
}

// ============================================================
// 场景 3: 异常路径 - 目标 IP 不可达
// ============================================================
func TestE2E_InvalidTargetIP(t *testing.T) {
	t.Log("=== 场景 3: 目标 IP 不可达 -> 任务失败 ===")

	// 1. 创建任务（目标 IP 不可达）
	taskID, err := createTaskWithTargetIP(t, "perf", "192.168.255.255", 1, 5, 99)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}
	t.Logf("✅ 任务创建成功: %s", taskID)

	// 2. 等待任务失败（最多 30 秒，因为 drop_server 可能更快返回错误）
	status, err := waitForTaskFailed(t, taskID, 30*time.Second)
	if err != nil {
		t.Fatalf("等待任务失败超时: %v", err)
	}
	t.Logf("✅ 任务状态: %s", status)
}

// ============================================================
// 辅助函数
// ============================================================

// TaskResponse API 返回的任务结构
type TaskResponse struct {
	ID             string `json:"id"`
	TaskID         string `json:"task_id"`
	TargetIP       string `json:"target_ip"`
	PID            int    `json:"pid"`
	Duration       int    `json:"duration"`
	Frequency      int    `json:"frequency"`
	ProfilerType   string `json:"profiler_type"`
	Status         string `json:"status"`
	StatusMsg      string `json:"status_msg"`
	StatusReason   string `json:"status_reason"`
	FlamegraphURL  string `json:"flamegraph_url"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
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

	// 解析 task_id
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
		if time.Since(start) > timeout {
			return "", fmt.Errorf("超时: 状态未变为 %s", targetStatus)
		}

		task, err := getTask(t, taskID)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		t.Logf("⏳ 当前状态: %s, 消息: %s", task.Status, task.StatusMsg)

		if task.Status == targetStatus {
			return task.Status, nil
		}
		if task.Status == stopStatus {
			return task.Status, fmt.Errorf("状态变为 %s 而不是 %s", stopStatus, targetStatus)
		}

		time.Sleep(2 * time.Second)
	}
}
