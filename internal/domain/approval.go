package domain

import "time"

type ApprovalStatus string

const (
	ApprovalStatusPending ApprovalStatus = "PENDING"
)

type Approval struct {
	ID              string
	ToolExecutionID string
	Status          ApprovalStatus
	RequestedBy     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
