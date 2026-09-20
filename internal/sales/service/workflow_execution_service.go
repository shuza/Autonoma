package service

import (
	"context"
	"fmt"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/sales/repository"
)

type WorkflowExecutionService struct {
	workflows     *repository.WorkflowRepository
	workflowSteps *repository.WorkflowStepRepository
}

func NewWorkflowExecutionService(workflows *repository.WorkflowRepository, workflowSteps *repository.WorkflowStepRepository) *WorkflowExecutionService {
	return &WorkflowExecutionService{
		workflows:     workflows,
		workflowSteps: workflowSteps,
	}
}

func (s *WorkflowExecutionService) ExecuteNextStep(ctx context.Context, workflow domain.Workflow) (ExecutionNextStepResult, error) {
	if s == nil || s.workflows == nil || s.workflowSteps == nil {
		return ExecutionNextStepResult{}, fmt.Errorf("workflow execution service is not configured")
	}

	if workflow.Status == domain.WorkflowStatusPending {
		if err := workflow.TransitionTo(domain.WorkflowStatusRunning); err != nil {
			return ExecutionNextStepResult{}, fmt.Errorf("failed to move workflow to running: %w", err)
		}

		var err error
		workflow, err = s.workflows.UpdateStatus(ctx, workflow.ID, workflow.Status)
		if err != nil {
			return ExecutionNextStepResult{}, fmt.Errorf("faile dto persis workflow status: %w", err)
		}
	} else if workflow.Status != domain.WorkflowStatusRunning {
		return ExecutionNextStepResult{}, fmt.Errorf("workflow status %s is not executable", workflow.Status)
	}

	steps, err := s.workflowSteps.ListByWorkflowID(ctx, workflow.ID)
	if err != nil {
		return ExecutionNextStepResult{}, fmt.Errorf("failed to load workflow steps: %w", err)
	}

	if hasRunningStep(steps) {
		for _, step := range steps {
			if step.Status == domain.WorkflowStepStatusRunning {
				return ExecutionNextStepResult{
					Workflow: workflow,
					Step:     step,
				}, nil
			}
		}
	}

	for _, step := range steps {
		if step.Status != domain.WorkflowStepStatusPending {
			continue
		}

		runningStep, err := s.workflowSteps.UpdateStatus(ctx, step.ID, domain.WorkflowStepStatusRunning)
		if err != nil {
			return ExecutionNextStepResult{}, fmt.Errorf("failed to mark workflow step running: %w", err)
		}

		return ExecutionNextStepResult{
			Workflow: workflow,
			Step:     runningStep,
		}, nil
	}

	return ExecutionNextStepResult{}, fmt.Errorf("no pending steps found")
}

func (s *WorkflowExecutionService) CancelWorkflow(ctx context.Context, workflow domain.Workflow) (domain.Workflow, error) {
	if s == nil || s.workflowSteps == nil {
		return domain.Workflow{}, fmt.Errorf("workflow execution service is not configured")
	}

	cancelledWorkflow, err := s.workflows.UpdateStatus(ctx, workflow.ID, domain.WorkflowStatusCancelled)
	if err != nil {
		return domain.Workflow{}, fmt.Errorf("failed to persist cancelled workflow status: %w", err)
	}

	return cancelledWorkflow, nil
}

func (s *WorkflowExecutionService) CompleteStep(ctx context.Context, workflow domain.Workflow, step domain.WorkflowStep) (ExecutionNextStepResult, error) {
	if s == nil || s.workflows == nil || s.workflowSteps == nil {
		return ExecutionNextStepResult{}, fmt.Errorf("workflow execution service is not configured")
	}

	if workflow.Status != domain.WorkflowStatusRunning {
		return ExecutionNextStepResult{}, fmt.Errorf("workflow status %s cannot complete a step", workflow.Status)
	}

	if step.Status != domain.WorkflowStepStatusRunning {
		return ExecutionNextStepResult{}, fmt.Errorf("workflow step status %s cannot be completed", step.Status)
	}

	completedStep, err := s.workflowSteps.UpdateStatus(ctx, step.ID, domain.WorkflowStepStatusCompleted)
	if err != nil {
		return ExecutionNextStepResult{}, fmt.Errorf("failed to mark workflow step completion: %w", err)
	}

	steps, err := s.workflowSteps.ListByWorkflowID(ctx, workflow.ID)
	if err != nil {
		return ExecutionNextStepResult{}, fmt.Errorf("failed to load workflow steps after completion: %w", err)
	}

	if hasPendingStep(steps) {
		return ExecutionNextStepResult{
			Workflow: workflow,
			Step:     completedStep,
		}, nil
	}

	if err := workflow.TransitionTo(domain.WorkflowStatusCompleted); err != nil {
		return ExecutionNextStepResult{}, fmt.Errorf("failed to complete workflow: %w", err)
	}

	updatedWorkflow, err := s.workflows.UpdateStatus(ctx, workflow.ID, domain.WorkflowStatusCompleted)
	if err != nil {
		return ExecutionNextStepResult{}, fmt.Errorf("failed to persist completed workflow status: %w", err)
	}

	return ExecutionNextStepResult{
		Workflow: updatedWorkflow,
		Step:     completedStep,
	}, nil
}

func (s *WorkflowExecutionService) FailStep(ctx context.Context, workflow domain.Workflow, step domain.WorkflowStep) (ExecutionNextStepResult, error) {
	if s == nil || s.workflows == nil || s.workflowSteps == nil {
		return ExecutionNextStepResult{}, fmt.Errorf("workflow execution service is not configured")
	}

	if workflow.Status != domain.WorkflowStatusRunning {
		return ExecutionNextStepResult{}, fmt.Errorf("workflow status %s cannot complete a step", workflow.Status)
	}

	if step.Status != domain.WorkflowStepStatusRunning {
		return ExecutionNextStepResult{}, fmt.Errorf("workflow step status %s cannot be completed", step.Status)
	}

	failedStep, err := s.workflowSteps.UpdateStatus(ctx, step.ID, domain.WorkflowStepStatusFailed)
	if err != nil {
		return ExecutionNextStepResult{}, fmt.Errorf("failed to mark workflow step failed: %w", err)
	}

	if err := workflow.TransitionTo(domain.WorkflowStatusFailed); err != nil {
		return ExecutionNextStepResult{}, fmt.Errorf("failed to move workflow to failed: %w", err)
	}

	failedWorkflow, err := s.workflows.UpdateStatus(ctx, workflow.ID, domain.WorkflowStatusFailed)
	if err != nil {
		return ExecutionNextStepResult{}, fmt.Errorf("failed to persist failed workflow status: %w", err)
	}

	return ExecutionNextStepResult{
		Workflow: failedWorkflow,
		Step:     failedStep,
	}, nil
}

func hasRunningStep(steps []domain.WorkflowStep) bool {
	for _, step := range steps {
		if step.Status == domain.WorkflowStepStatusRunning {
			return true
		}
	}
	return false
}

func hasPendingStep(steps []domain.WorkflowStep) bool {
	for _, step := range steps {
		if step.Status == domain.WorkflowStepStatusPending {
			return true
		}
	}
	return false
}

type ExecutionNextStepResult struct {
	Workflow domain.Workflow
	Step     domain.WorkflowStep
}
