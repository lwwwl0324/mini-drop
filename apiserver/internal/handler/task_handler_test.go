package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"mini-drop/apiserver/internal/model"
	"mini-drop/apiserver/internal/service"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("无法创建测试数据库: %v", err)
	}
	db.AutoMigrate(&model.Task{})
	return db
}

func setupTestHandler(t *testing.T) (*TaskHandler, *gorm.DB) {
	db := setupTestDB(t)

	taskService := &service.TaskService{
		Db: db,
	}

	handler := &TaskHandler{
		service: taskService,
	}

	return handler, db
}

func TestCreateTask_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _ := setupTestHandler(t)

	testCases := []struct {
		name       string
		body       map[string]interface{}
		expectCode int
	}{
		{
			name: "缺少 target_ip",
			body: map[string]interface{}{
				"pid":       1234,
				"duration":  10,
				"frequency": 999,
			},
			expectCode: 400,
		},
		{
			name: "缺少 duration",
			body: map[string]interface{}{
				"target_ip": "127.0.0.1",
				"pid":       1234,
				"frequency": 999,
			},
			expectCode: 400,
		},
		{
			name: "缺少 frequency",
			body: map[string]interface{}{
				"target_ip": "127.0.0.1",
				"pid":       1234,
				"duration":  10,
			},
			expectCode: 400,
		},
		{
			name: "PID 为负数",
			body: map[string]interface{}{
				"target_ip": "127.0.0.1",
				"pid":       -1,
				"duration":  10,
				"frequency": 999,
			},
			expectCode: 400,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tc.body)

			req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.CreateTask(c)

			if w.Code != tc.expectCode {
				t.Errorf("期望状态码 %d, 实际=%d", tc.expectCode, w.Code)
			}
		})
	}
}

func TestGetTask_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/tasks/nonexistent", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "nonexistent"}}

	handler.GetTask(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("期望状态码 404, 实际=%d", w.Code)
	}
}

func TestGetTask_WithValidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db := setupTestHandler(t)

	task := &model.Task{
		TaskID:   "valid_task_001",
		TargetIP: "127.0.0.1",
		PID:      1234,
		Duration: 10,
		Frequency: 999,
		Status:   string(model.StatusDone),
	}
	db.Create(task)

	req := httptest.NewRequest("GET", "/api/v1/tasks/valid_task_001", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "valid_task_001"}}

	handler.GetTask(c)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200, 实际=%d", w.Code)
	}
}

func TestListTasks_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.ListTasks(c)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200, 实际=%d", w.Code)
	}
}

func TestListTasks_WithData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db := setupTestHandler(t)

	db.Create(&model.Task{
		TaskID:   "task_001",
		TargetIP: "127.0.0.1",
		PID:      1234,
		Duration: 10,
		Frequency: 999,
		Status:   string(model.StatusDone),
	})

	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.ListTasks(c)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200, 实际=%d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["code"] != float64(0) {
		t.Errorf("期望 code=0, 实际=%v", resp["code"])
	}
}
