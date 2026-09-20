package domain

import "time"

type WorkflowStepStatus string

const (
	WorkflowStepStatusPending   WorkflowStepStatus = "PENDING"
	WorkflowStepStatusRunning   WorkflowStepStatus = "RUNNING"
	WorkflowStepStatusCompleted WorkflowStepStatus = "COMPLETED"
)

type WorkflowStep struct {
	ID         string
	WorkflowID string
	Name       string
	Status     WorkflowStepStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
