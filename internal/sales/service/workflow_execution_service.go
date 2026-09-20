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

	for _, step := range steps {
		if step.Status == domain.WorkflowStepStatusRunning {
			return ExecutionNextStepResult{
				Workflow: workflow,
				Step:     step,
			}, nil
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

type ExecutionNextStepResult struct {
	Workflow domain.Workflow
	Step     domain.WorkflowStep
}
