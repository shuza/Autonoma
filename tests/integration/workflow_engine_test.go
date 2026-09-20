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

func TestWorkflowResumeReturnsExistingRunningStep(t *testing.T) {
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
		t.Fatalf("expected updated workflow status %q, got %q", domain.WorkflowStatusRunning, updatedWorkflow.Status)
	}

	if updatedStep.Status != domain.WorkflowStepStatusRunning {
		t.Fatalf("expected updated workflow step status %q, got %q", domain.WorkflowStepStatusRunning, updatedStep.Status)
	}

	runningSteps, err := fixture.workflowStepRepository.ListByWorkflowID(ctx, fixture.workflow.ID)
	if err != nil {
		t.Fatalf("failed to load workflow steps during resume: %v", err)
	}

	if len(runningSteps) != 1 || runningSteps[0].Status != domain.WorkflowStepStatusRunning {
		t.Fatalf("unexpected workflow steps during resume: %+v", runningSteps)
	}
}

func TestWorkflowCancellationIsPersisted(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)

	cancelledWorkflow, err := fixture.workflowRepository.UpdateStatus(ctx, fixture.workflow.ID, domain.WorkflowStatusCancelled)
	if err != nil {
		t.Fatalf("failed to cancel workflow: %v", err)
	}

	if cancelledWorkflow.Status != domain.WorkflowStatusCancelled {
		t.Fatalf("expected cancelled workflow status %q, got %q", domain.WorkflowStatusCancelled, cancelledWorkflow.Status)
	}

	workflows, err := fixture.workflowRepository.ListByLeadID(ctx, fixture.lead.ID)
	if err != nil {
		t.Fatalf("failed to reload workflows after cancellation: %v", err)
	}

	if len(workflows) != 1 || workflows[0].Status != domain.WorkflowStatusCancelled {
		t.Fatalf("unexpected workflow state after cancellation: %+v", workflows)
	}
}

func TestWorkflowCompletionIsPersisted(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)

	runningWorkflow, err := fixture.workflowRepository.UpdateStatus(ctx, fixture.workflow.ID, domain.WorkflowStatusRunning)
	if err != nil {
		t.Fatalf("failed to update workflow status to running: %v", err)
	}

	completedStep, err := fixture.workflowStepRepository.UpdateStatus(ctx, fixture.workflowStep.ID, domain.WorkflowStepStatusCompleted)
	if err != nil {
		t.Fatalf("failed to update workflow step status to completed: %v", err)
	}

	completedWorkflow, err := fixture.workflowRepository.UpdateStatus(ctx, runningWorkflow.ID, domain.WorkflowStatusCompleted)
	if err != nil {
		t.Fatalf("failed to update workflow status to completed: %v", err)
	}

	if completedStep.Status != domain.WorkflowStepStatusCompleted {
		t.Fatalf("expected workflow step status %q, got %q", domain.WorkflowStepStatusCompleted, completedStep.Status)
	}

	if completedWorkflow.Status != domain.WorkflowStatusCompleted {
		t.Fatalf("expected workflow status %q, got %q", domain.WorkflowStatusCompleted, completedWorkflow.Status)
	}
}

func TestWorkflowFailureIsPersisted(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)

	runningWorkflow, err := fixture.workflowRepository.UpdateStatus(ctx, fixture.workflow.ID, domain.WorkflowStatusRunning)
	if err != nil {
		t.Fatalf("failed to update workflow status to running: %v", err)
	}

	failedStep, err := fixture.workflowStepRepository.UpdateStatus(ctx, fixture.workflowStep.ID, domain.WorkflowStepStatusFailed)
	if err != nil {
		t.Fatalf("failed to update workflow step status to failed: %v", err)
	}

	failedWorkflow, err := fixture.workflowRepository.UpdateStatus(ctx, runningWorkflow.ID, domain.WorkflowStatusFailed)
	if err != nil {
		t.Fatalf("failed to update workflow status to failed: %v", err)
	}

	if failedStep.Status != domain.WorkflowStepStatusFailed {
		t.Fatalf("expected workflow step status %q, got %q", domain.WorkflowStepStatusFailed, failedStep.Status)
	}

	if failedWorkflow.Status != domain.WorkflowStatusFailed {
		t.Fatalf("expected workflow status %q, got %q", domain.WorkflowStatusFailed, failedWorkflow.Status)
	}
}

func TestWorkflowRetryIsPersisted(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)

	failedWorkflow, err := fixture.workflowRepository.UpdateStatus(ctx, fixture.workflow.ID, domain.WorkflowStatusFailed)
	if err != nil {
		t.Fatalf("failed to update workflow status to failed: %v", err)
	}

	failedStep, err := fixture.workflowStepRepository.UpdateStatus(ctx, fixture.workflowStep.ID, domain.WorkflowStepStatusFailed)
	if err != nil {
		t.Fatalf("failed to update workflow step status to failed: %v", err)
	}

	retriedWorkflow, err := fixture.workflowRepository.UpdateStatus(ctx, failedWorkflow.ID, domain.WorkflowStatusRunning)
	if err != nil {
		t.Fatalf("failed to update workflow status to running: %v", err)
	}

	retriedStep, err := fixture.workflowStepRepository.UpdateStatus(ctx, failedStep.ID, domain.WorkflowStepStatusRunning)
	if err != nil {
		t.Fatalf("failed to update workflow step status to running: %v", err)
	}

	if retriedWorkflow.Status != domain.WorkflowStatusRunning {
		t.Fatalf("expected workflow status %q, got %q", domain.WorkflowStatusRunning, retriedWorkflow.Status)
	}

	if retriedStep.Status != domain.WorkflowStepStatusRunning {
		t.Fatalf("expected workflow step status %q, got %q", domain.WorkflowStepStatusRunning, retriedStep.Status)
	}
}
