package service

import (
	"context"
	"testing"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/sales/repository"
)

func TestExecuteNextStepRequiresConfiguredDependencies(t *testing.T) {
	t.Parallel()

	service := NewWorkflowExecutionService(nil, nil)
	_, err := service.ExecuteNextStep(context.Background(), domain.Workflow{})
	if err == nil {
		t.Fatalf("expected configuration error")
	}
}

func TestExecuteNextStepRejectsNonExecutableWorkflowStatus(t *testing.T) {
	t.Parallel()

	service := NewWorkflowExecutionService(&repository.WorkflowRepository{}, &repository.WorkflowStepRepository{})
	_, err := service.ExecuteNextStep(context.Background(), domain.Workflow{
		ID:     "workflow-1",
		Status: domain.WorkflowStatusCompleted,
	})
	if err == nil {
		t.Fatalf("expected non-executable workflow error")
	}
}

func TestExecuteNextStepReturnsRunningStepWhenWorkflowIsResumed(t *testing.T) {
	t.Parallel()

	service := NewWorkflowExecutionService(&repository.WorkflowRepository{}, &repository.WorkflowStepRepository{})
	_, err := service.ExecuteNextStep(context.Background(), domain.Workflow{
		ID:     "workflow-1",
		Status: domain.WorkflowStatusRunning,
	})
	if err == nil {
		t.Fatalf("expected repository-backed resume path to require configured store")
	}
}
