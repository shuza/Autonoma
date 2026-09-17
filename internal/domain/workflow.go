package domain

import "time"

type WorkflowStatus string

const (
	WorkflowStatusNew WorkflowStatus = "NEW"
)

type Workflow struct {
	ID        string
	LeadID    string
	Status    WorkflowStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
