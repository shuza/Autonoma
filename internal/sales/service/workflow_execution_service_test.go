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

func TestCancelWorkflowRejectsNonCancellableStatus(t *testing.T) {
	t.Parallel()

	service := NewWorkflowExecutionService(&repository.WorkflowRepository{}, &repository.WorkflowStepRepository{})
	_, err := service.CancelWorkflow(context.Background(), domain.Workflow{
		ID:     "workflow-1",
		Status: domain.WorkflowStatusCompleted,
	})
	if err == nil {
		t.Fatalf("expected cancellation to fail for completed workflow")
	}
}

func TestCompleteStepRejectsNonRunningWorkflow(t *testing.T) {
	t.Parallel()

	service := NewWorkflowExecutionService(&repository.WorkflowRepository{}, &repository.WorkflowStepRepository{})
	_, err := service.CompleteStep(context.Background(), domain.Workflow{
		ID:     "workflow-1",
		Status: domain.WorkflowStatusPending,
	}, domain.WorkflowStep{
		ID:     "step-1",
		Status: domain.WorkflowStepStatusRunning,
	})
	if err == nil {
		t.Fatalf("expected completion to fail for non-running workflow")
	}
}

func TestFailStepRejectNonRunningStep(t *testing.T) {
	t.Parallel()

	service := NewWorkflowExecutionService(&repository.WorkflowRepository{}, &repository.WorkflowStepRepository{})
	_, err := service.FailStep(context.Background(), domain.Workflow{
		ID:     "workflow-1",
		Status: domain.WorkflowStatusRunning,
	}, domain.WorkflowStep{
		ID:     "step-1",
		Status: domain.WorkflowStepStatusPending,
	})
	if err == nil {
		t.Fatalf("expected failing a non-running step to fail")
	}
}
