package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"mini-drop/apiserver/internal/model"
	"mini-drop/apiserver/internal/service"
)

type HeartbeatHandler struct {
	service *service.AgentService
	db      *gorm.DB
}

func NewHeartbeatHandler(service *service.AgentService, db *gorm.DB) *HeartbeatHandler {
	return &HeartbeatHandler{service: service, db: db}
}

type HeartbeatRequest struct {
	AgentID  string `json:"agent_id"`
	Hostname string `json:"hostname"`
	IPAddr   string `json:"ip_addr"`
	Version  string `json:"version"`
}

func (h *HeartbeatHandler) Receive(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	// 更新或创建 Agent
	var agent model.Agent
	result := h.db.Where("agent_id = ?", req.AgentID).First(&agent)
	
	if result.Error != nil {
		// 新 Agent，创建并记录上线日志
		agent = model.Agent{
			AgentID:       req.AgentID,
			Hostname:      req.Hostname,
			IPAddr:        req.IPAddr,
			Version:       req.Version,
			Status:        "online",
			LastHeartbeat: time.Now(),
		}
		h.db.Create(&agent)
		
		// 记录审计日志
		h.createAuditLog(req.AgentID, "online", "Agent 首次上线", req.IPAddr)
	} else {
		oldStatus := agent.Status
		agent.Hostname = req.Hostname
		agent.IPAddr = req.IPAddr
		agent.Version = req.Version
		agent.LastHeartbeat = time.Now()
		
		// 如果之前是离线，现在上线了
		if oldStatus == "offline" {
			agent.Status = "online"
			h.createAuditLog(req.AgentID, "online", "Agent 恢复上线", req.IPAddr)
		}
		
		h.db.Save(&agent)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "heartbeat received",
	})
}

func (h *HeartbeatHandler) createAuditLog(agentID, eventType, detail, ipAddr string) {
	log := model.AuditLog{
		AgentID:     agentID,
		EventType:   eventType,
		EventDetail: detail,
		IPAddr:      ipAddr,
		CreatedAt:   time.Now(),
	}
	h.db.Create(&log)
}
