package domain

import "time"

type ToolExecutionStatus string

const (
	ToolExecutionStatusPending ToolExecutionStatus = "PENDING"
)

type ToolExecution struct {
	ID             string
	WorkflowStepID string
	ToolName       string
	Status         ToolExecutionStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
