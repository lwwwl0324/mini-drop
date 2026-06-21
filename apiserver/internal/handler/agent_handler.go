package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mini-drop/apiserver/internal/service"
)

type AgentHandler struct {
	service *service.AgentService
}

func NewAgentHandler(service *service.AgentService) *AgentHandler {
	return &AgentHandler{service: service}
}

func (h *AgentHandler) ListAgents(c *gin.Context) {
	agents, err := h.service.ListAgents()
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
		"data": agents,
	})
}

func (h *AgentHandler) GetAuditLogs(c *gin.Context) {
	agentID := c.Param("agent_id")
	logs, err := h.service.GetAuditLogs(agentID)
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
		"data": logs,
	})
}
