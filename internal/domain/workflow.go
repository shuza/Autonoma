package domain

import (
	"fmt"
	"time"
)

type WorkflowStatus string

const (
	WorkflowStatusPending         WorkflowStatus = "PENDING"
	WorkflowStatusRunning         WorkflowStatus = "RUNNING"
	WorkflowStatusWaitingApproval WorkflowStatus = "WAITING_APPROVAL"
	WorkflowStatusCompleted       WorkflowStatus = "COMPLETED"
	WorkflowStatusFailed          WorkflowStatus = "FAILED"
	WorkflowStatusCancelled       WorkflowStatus = "CANCELLED"
)

type Workflow struct {
	ID        string
	LeadID    string
	Status    WorkflowStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

var allowedWorkflowTransitions = map[WorkflowStatus]map[WorkflowStatus]struct{}{
	WorkflowStatusPending: {
		WorkflowStatusRunning:   {},
		WorkflowStatusCancelled: {},
	},
	WorkflowStatusRunning: {
		WorkflowStatusWaitingApproval: {},
		WorkflowStatusCompleted:       {},
		WorkflowStatusFailed:          {},
		WorkflowStatusCancelled:       {},
	},
	WorkflowStatusWaitingApproval: {
		WorkflowStatusRunning:   {},
		WorkflowStatusCancelled: {},
	},
	WorkflowStatusCompleted: {},
	WorkflowStatusFailed:    {},
	WorkflowStatusCancelled: {},
}

func (w Workflow) CanTransitionTo(next WorkflowStatus) bool {
	allowedNext, ok := allowedWorkflowTransitions[w.Status]
	if !ok {
		return false
	}

	_, ok = allowedNext[next]
	return ok
}

func (w *Workflow) TransitionTo(next WorkflowStatus) error {
	if w == nil {
		return fmt.Errorf("workflow is required")
	}

	if !w.CanTransitionTo(next) {
		return fmt.Errorf("invalid workflow transition from %s to %s", w.Status, next)
	}

	w.Status = next
	return nil
}
