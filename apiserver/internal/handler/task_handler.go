package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mini-drop/apiserver/internal/service"
)

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

type CreateTaskRequest struct {
	TargetIP     string `json:"target_ip" binding:"required"`
	PID          int    `json:"pid"`
	Duration     int    `json:"duration" binding:"required"`
	Frequency    int    `json:"frequency" binding:"required"`
	ProfilerType string `json:"profiler_type"` // perf, ebpf, pyspy
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	if req.PID < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "PID 不能为负数",
		})
		return
	}

	if req.ProfilerType == "" {
		req.ProfilerType = "perf"
	}

	// 将字符串转换为数字
	var profilerTypeInt int
	switch req.ProfilerType {
	case "ebpf":
		profilerTypeInt = 1
	case "pyspy":
		profilerTypeInt = 2
	default:
		profilerTypeInt = 0
	}

	println("=== DEBUG ===")
	println("req.ProfilerType:", req.ProfilerType)
	println("profilerTypeInt:", profilerTypeInt)
	println("=== END DEBUG ===")

	task, err := h.service.CreateTask(c.Request.Context(), req.TargetIP, req.PID, req.Duration, req.Frequency, profilerTypeInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": task,
	})
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	task, err := h.service.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "任务不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": task,
	})
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	tasks, err := h.service.ListTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": tasks,
	})
}
