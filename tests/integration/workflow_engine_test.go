package integration

import (
	"testing"

	"github.com/shuza/Autonoma/internal/domain"
)

func TestWorkflowStatusTransitionIsPersisted(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)

	if err := fixture.workflow.TransitionTo(domain.WorkflowStatusRunning); err != nil {
		t.Fatalf("failed to transition workflow status: %v", err)
	}

	updatedWorkflow, err := fixture.workflowRepository.UpdateStatus(ctx, fixture.workflow.ID, fixture.workflow.Status)
	if err != nil {
		t.Fatalf("failed to update workflow status: %v", err)
	}

	if updatedWorkflow.Status != domain.WorkflowStatusRunning {
		t.Fatalf("expected updated workflow status %q, got %q", domain.WorkflowStatusRunning, updatedWorkflow.Status)
	}

	workflows, err := fixture.workflowRepository.ListByLeadID(ctx, fixture.lead.ID)
	if err != nil {
		t.Fatalf("failed to reload workflows after status update: %v", err)
	}

	if len(workflows) != 1 || workflows[0].Status != domain.WorkflowStatusRunning {
		t.Fatalf("unexpected workflow status after status update: %+v", workflows)
	}
}

func TestWorkflowStepExecutionPersistsRunningState(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)

	updatedWorkflow, err := fixture.workflowRepository.UpdateStatus(ctx, fixture.workflow.ID, domain.WorkflowStatusRunning)
	if err != nil {
		t.Fatalf("failed to update workflow status to running: %v", err)
	}

	updatedStep, err := fixture.workflowStepRepository.UpdateStatus(ctx, fixture.workflowStep.ID, domain.WorkflowStepStatusRunning)
	if err != nil {
		t.Fatalf("failed to update workflow step status to running: %v", err)
	}

	if updatedWorkflow.Status != domain.WorkflowStatusRunning {
		t.Fatalf("expected updated workflow step status %q, got %q", domain.WorkflowStepStatusRunning, updatedStep.Status)
	}

	if updatedStep.Status != domain.WorkflowStepStatusRunning {
		t.Fatalf("expected updated workflow step status %q, got %q", domain.WorkflowStepStatusRunning, updatedStep.Status)
	}

	steps, err := fixture.workflowStepRepository.ListByWorkflowID(ctx, fixture.workflow.ID)
	if err != nil {
		t.Fatalf("failed to reload workflow steps after status update: %v", err)
	}

	if len(steps) != 1 || steps[0].Status != domain.WorkflowStepStatusRunning {
		t.Fatalf("unexpected workflow step status after status update: %+v", steps)
	}
}
