package model

import "testing"

func TestAgentTableName(t *testing.T) {
	agent := Agent{}
	if agent.TableName() != "agents" {
		t.Errorf("期望表名 'agents', 实际 '%s'", agent.TableName())
	}
}

func TestAuditLogTableName(t *testing.T) {
	log := AuditLog{}
	if log.TableName() != "audit_logs" {
		t.Errorf("期望表名 'audit_logs', 实际 '%s'", log.TableName())
	}
}
