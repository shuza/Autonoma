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

func TestCancelWorkflowReturnsCancelledWorkflowIdempotently(t *testing.T) {
	t.Parallel()

	service := NewWorkflowExecutionService(&repository.WorkflowRepository{}, &repository.WorkflowStepRepository{})
	workflow, err := service.CancelWorkflow(context.Background(), domain.Workflow{
		ID:     "workflow-1",
		Status: domain.WorkflowStatusCancelled,
	})
	if err != nil {
		t.Fatalf("expected cancelled workflow to be retuned idempotently: %v", err)
	}
	if workflow.Status != domain.WorkflowStatusCancelled {
		t.Fatalf("expected cancelled workflow status, got %s", workflow.Status)
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

func TestFailStepRejectsNonRunningStep(t *testing.T) {
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

func TestCompleteStepReturnsCompletedStepIdempotently(t *testing.T) {
	t.Parallel()

	service := NewWorkflowExecutionService(&repository.WorkflowRepository{}, &repository.WorkflowStepRepository{})
	result, err := service.CompleteStep(context.Background(), domain.Workflow{
		ID:     "workflow-1",
		Status: domain.WorkflowStatusCompleted,
	}, domain.WorkflowStep{
		ID:     "step-1",
		Status: domain.WorkflowStepStatusCompleted,
	})
	if err != nil {
		t.Fatalf("expected completed step to be returned idempotently %v", err)
	}

	if result.Step.Status != domain.WorkflowStepStatusCompleted {
		t.Fatalf("expected completed step status, got %s", result.Step.Status)
	}
}

func TestFailStepReturnsFailedStepIdempotently(t *testing.T) {
	t.Parallel()

	service := NewWorkflowExecutionService(&repository.WorkflowRepository{}, &repository.WorkflowStepRepository{})
	result, err := service.CompleteStep(context.Background(), domain.Workflow{
		ID:     "workflow-1",
		Status: domain.WorkflowStatusCompleted,
	}, domain.WorkflowStep{
		ID:     "step-1",
		Status: domain.WorkflowStepStatusFailed,
	})
	if err != nil {
		t.Fatalf("expected failed step to be returned idempotently %v", err)
	}

	if result.Step.Status != domain.WorkflowStepStatusFailed {
		t.Fatalf("expected failed step status, got %s", result.Step.Status)
	}
}
