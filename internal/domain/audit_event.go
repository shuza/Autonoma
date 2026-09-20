package domain

import "time"

type AuditEvent struct {
	ID         string
	WorkflowID string
	EventType  string
	Actor      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
