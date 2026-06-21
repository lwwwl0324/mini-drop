package service

import (
	"time"

	"gorm.io/gorm"

	"mini-drop/apiserver/internal/model"
)

type AgentService struct {
	db *gorm.DB
}

func NewAgentService(db *gorm.DB) *AgentService {
	return &AgentService{db: db}
}

type AgentInfo struct {
	UID           string    `json:"uid"`
	Hostname      string    `json:"hostname"`
	IPAddr        string    `json:"ip_addr"`
	Version       string    `json:"version"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	Status        string    `json:"status"`
}

// CreateAuditLog 公开方法，用于创建审计日志
func (s *AgentService) CreateAuditLog(agentID, eventType, detail, ipAddr string) {
	log := model.AuditLog{
		AgentID:     agentID,
		EventType:   eventType,
		EventDetail: detail,
		IPAddr:      ipAddr,
		CreatedAt:   time.Now(),
	}
	s.db.Create(&log)
}

func (s *AgentService) ListAgents() ([]AgentInfo, error) {
	var agents []model.Agent
	if err := s.db.Find(&agents).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	for i := range agents {
		if agents[i].Status == "online" {
			elapsed := now.Sub(agents[i].LastHeartbeat).Seconds()
			if elapsed > 30 {
				agents[i].Status = "offline"
				s.db.Save(&agents[i])
				s.CreateAuditLog(agents[i].AgentID, "offline", "Agent 离线 (心跳超时 30 秒)", agents[i].IPAddr)
			}
		}
	}

	result := make([]AgentInfo, len(agents))
	for i, a := range agents {
		result[i] = AgentInfo{
			UID:           a.AgentID,
			Hostname:      a.Hostname,
			IPAddr:        a.IPAddr,
			Version:       a.Version,
			LastHeartbeat: a.LastHeartbeat,
			Status:        a.Status,
		}
	}
	return result, nil
}

func (s *AgentService) UpsertAgent(agentID, hostname, ipAddr, version, status string) error {
	agent := model.Agent{
		AgentID:       agentID,
		Hostname:      hostname,
		IPAddr:        ipAddr,
		Version:       version,
		Status:        status,
		LastHeartbeat: time.Now(),
	}
	return s.db.Where("agent_id = ?", agentID).Assign(agent).FirstOrCreate(&agent).Error
}

func (s *AgentService) GetAuditLogs(agentID string) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := s.db.Where("agent_id = ?", agentID).Order("created_at DESC").Limit(100).Find(&logs).Error
	return logs, err
}
