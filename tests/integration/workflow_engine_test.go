package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/sales/service"
)

func TestExecuteNextStepPersistsRunningWorkflowAndStep(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)
	executionService := service.NewWorkflowExecutionService(fixture.workflowRepository, fixture.workflowStepRepository)

	result, err := executionService.ExecuteNextStep(ctx, fixture.workflow)
	if err != nil {
		t.Fatalf("failed to execute next step: %v", err)
	}

	if result.Workflow.Status != domain.WorkflowStatusRunning {
		t.Fatalf("expected workflow status %q, got %q", domain.WorkflowStatusRunning, result.Workflow.Status)
	}

	if result.Step.Status != domain.WorkflowStepStatusRunning {
		t.Fatalf("expected step status %q, got %q", domain.WorkflowStepStatusRunning, result.Step.Status)
	}

	workflows, err := fixture.workflowRepository.ListByLeadID(ctx, fixture.lead.ID)
	if err != nil {
		t.Fatalf("failed to reload workflows after execution: %v", err)
	}

	if len(workflows) != 1 || workflows[0].Status != domain.WorkflowStatusRunning {
		t.Fatalf("unexpected workflow status after status update: %+v", workflows)
	}

	steps, err := fixture.workflowStepRepository.ListByWorkflowID(ctx, fixture.workflow.ID)
	if err != nil {
		t.Fatalf("failed to reload workflow steps after execution: %v", err)
	}

	if len(steps) != 1 || steps[0].Status != domain.WorkflowStepStatusRunning {
		t.Fatalf("unexpected workflow step status after execution: %+v", steps)
	}
}

func TestWorkflowResumeReturnsExistingRunningStepAfterRestart(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	firstStore := newIntegrationStore(t, ctx, databaseURL)

	fixture := createPersistenceFixture(t, ctx, firstStore)
	executionService := service.NewWorkflowExecutionService(fixture.workflowRepository, fixture.workflowStepRepository)

	firstResult, err := executionService.ExecuteNextStep(ctx, fixture.workflow)
	if err != nil {
		t.Fatalf("failed to start workflow execution: %v", err)
	}
	firstStore.Close()

	restartedStore := newIntegrationStore(t, ctx, databaseURL)
	defer restartedStore.Close()

	reloadedWorkflows, err := fixture.newWorkflowRepository(restartedStore).ListByLeadID(ctx, fixture.lead.ID)
	if err != nil {
		t.Fatalf("failed to reload workflows after restart: %v", err)
	}
	if len(reloadedWorkflows) != 1 {
		t.Fatalf("expected 1 workflow after restart, got %d", len(reloadedWorkflows))
	}

	restartedService := service.NewWorkflowExecutionService(
		fixture.newWorkflowRepository(restartedStore),
		fixture.newWorkflowStepRepository(restartedStore),
	)
	resumedResult, err := restartedService.ExecuteNextStep(ctx, reloadedWorkflows[0])
	if err != nil {
		t.Fatalf("failed to resume workflow after restart: %v", err)
	}

	if resumedResult.Workflow.Status != domain.WorkflowStatusRunning {
		t.Fatalf("expected resumed workflow status %q, got %q", domain.WorkflowStatusRunning, resumedResult.Workflow.Status)
	}
	if resumedResult.Step.ID != firstResult.Step.ID {
		t.Fatalf("expected resumed step %q, got %q", firstResult.Step.ID, resumedResult.Step.ID)
	}
	if resumedResult.Step.Status != domain.WorkflowStepStatusRunning {
		t.Fatalf("expected resumed step status %q, got %q", domain.WorkflowStepStatusRunning, resumedResult.Step.Status)
	}
}

func TestPausedForApprovalPersistWaitingApprovalState(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)
	executionService := service.NewWorkflowExecutionService(fixture.workflowRepository, fixture.workflowStepRepository)

	runningResult, err := executionService.ExecuteNextStep(ctx, fixture.workflow)
	if err != nil {
		t.Fatalf("failed to execute workflow step: %v", err)
	}

	waitingResult, err := executionService.PauseForApproval(ctx, runningResult.Workflow, runningResult.Step)
	if err != nil {
		t.Fatalf("failed to pause workflow for approval: %v", err)
	}

	if waitingResult.Workflow.Status != domain.WorkflowStatusWaitingApproval {
		t.Fatalf("expected workflow status %q, got %q", domain.WorkflowStatusWaitingApproval, waitingResult.Workflow.Status)
	}
	if waitingResult.Step.Status != domain.WorkflowStepStatusRunning {
		t.Fatalf("expected workflow step status %q, got %q", domain.WorkflowStepStatusRunning, waitingResult.Step.Status)
	}
}

func TestResumeAfterApprovalPersistsRunningState(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)
	executionService := service.NewWorkflowExecutionService(fixture.workflowRepository, fixture.workflowStepRepository)

	runningResult, err := executionService.ExecuteNextStep(ctx, fixture.workflow)
	if err != nil {
		t.Fatalf("failed to execute workflow step: %v", err)
	}

	waitingResult, err := executionService.PauseForApproval(ctx, runningResult.Workflow, runningResult.Step)
	if err != nil {
		t.Fatalf("failed to pause workflow for approval: %v", err)
	}

	resumedResult, err := executionService.ResumeAfterApproval(ctx, waitingResult.Workflow, waitingResult.Step)
	if err != nil {
		t.Fatalf("failed to resume workflow after approval: %v", err)
	}

	if resumedResult.Workflow.Status != domain.WorkflowStatusRunning {
		t.Fatalf("expected workflow status %q, got %v", domain.WorkflowStatusRunning, resumedResult.Workflow.Status)
	}
	if resumedResult.Step.Status != domain.WorkflowStepStatusRunning {
		t.Fatalf("expected workflow step status %q, got %v", domain.WorkflowStepStatusRunning, resumedResult.Step.Status)
	}
}

func TestWorkflowCancellationIsPersistedThroughServicee(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)
	executionService := service.NewWorkflowExecutionService(fixture.workflowRepository, fixture.workflowStepRepository)

	cancelledWorkflow, err := executionService.CancelWorkflow(ctx, fixture.workflow)
	if err != nil {
		t.Fatalf("failed to cancel workflow: %v", err)
	}

	if cancelledWorkflow.Status != domain.WorkflowStatusCancelled {
		t.Fatalf("expected workflow status %q, got %v", domain.WorkflowStatusCancelled, cancelledWorkflow.Status)
	}

	workflows, err := fixture.workflowRepository.ListByLeadID(ctx, fixture.lead.ID)
	if err != nil {
		t.Fatalf("failed to reload workflows after cancellation: %v", err)
	}

	if len(workflows) != 1 || workflows[0].Status != domain.WorkflowStatusCancelled {
		t.Fatalf("unexpected workflow state after cancellation: %v", workflows[0].Status)
	}
}

func TestWorkflowCompletionIsPersistedThroughService(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)
	executionService := service.NewWorkflowExecutionService(fixture.workflowRepository, fixture.workflowStepRepository)

	runningResult, err := executionService.ExecuteNextStep(ctx, fixture.workflow)
	if err != nil {
		t.Fatalf("failed to execute workflow step: %v", err)
	}

	completedResult, err := executionService.CompleteStep(ctx, runningResult.Workflow, runningResult.Step)
	if err != nil {
		t.Fatalf("failed to complete workflow step: %v", err)
	}

	if completedResult.Step.Status != domain.WorkflowStepStatusCompleted {
		t.Fatalf("expected workflow step status %q, got %q", domain.WorkflowStepStatusCompleted, completedResult.Step.Status)
	}
	if completedResult.Workflow.Status != domain.WorkflowStatusCompleted {
		t.Fatalf("expected workflow status %q, got %q", domain.WorkflowStatusCompleted, completedResult.Workflow.Status)
	}
}

func TestWorkflowFailureIsPersistedThroughService(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)
	executionService := service.NewWorkflowExecutionService(fixture.workflowRepository, fixture.workflowStepRepository)

	runningResult, err := executionService.ExecuteNextStep(ctx, fixture.workflow)
	if err != nil {
		t.Fatalf("failed to execute workflow step: %v", err)
	}

	failedResult, err := executionService.FailStep(ctx, runningResult.Workflow, runningResult.Step)
	if err != nil {
		t.Fatalf("failed to fail workflow step: %v", err)
	}

	if failedResult.Step.Status != domain.WorkflowStepStatusFailed {
		t.Fatalf("expected workflow step status %q, got %v", domain.WorkflowStepStatusFailed, failedResult.Step.Status)
	}
	if failedResult.Workflow.Status != domain.WorkflowStatusFailed {
		t.Fatalf("expected workflow status %q, got %v", domain.WorkflowStatusFailed, failedResult.Workflow.Status)
	}
}

func TestWorkflowRetryRestartsFailedStepThroughService(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)
	executionService := service.NewWorkflowExecutionService(fixture.workflowRepository, fixture.workflowStepRepository)

	runningResult, err := executionService.ExecuteNextStep(ctx, fixture.workflow)
	if err != nil {
		t.Fatalf("failed to execute workflow step: %v", err)
	}

	failedResult, err := executionService.FailStep(ctx, runningResult.Workflow, runningResult.Step)
	if err != nil {
		t.Fatalf("failed to fail workflow step: %v", err)
	}

	retriedResult, err := executionService.RetryStep(ctx, failedResult.Workflow, failedResult.Step)
	if err != nil {
		t.Fatalf("failed to retry workflow step: %v", err)
	}

	if retriedResult.Workflow.Status != domain.WorkflowStatusRunning {
		t.Fatalf("expected workflow status %q, got %q", domain.WorkflowStatusRunning, retriedResult.Workflow.Status)
	}
	if retriedResult.Step.Status != domain.WorkflowStepStatusRunning {
		t.Fatalf("expected workflow step status %q, got %q", domain.WorkflowStepStatusRunning, retriedResult.Step.Status)
	}
}

func TestWorkflowRetryPolicyRejectsExceededAttempts(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)
	executionService := service.NewWorkflowExecutionService(fixture.workflowRepository, fixture.workflowStepRepository)

	_, err := fixture.workflowRepository.UpdateStatus(ctx, fixture.workflow.ID, domain.WorkflowStatusFailed)
	if err != nil {
		t.Fatalf("failed to mark workflow status to failed: %v", err)
	}

	for i := 0; i < 2; i++ {
		_, err := fixture.workflowStepRepository.Create(ctx, domain.WorkflowStep{
			ID:         uuid.NewString(),
			WorkflowID: fixture.workflow.ID,
			Name:       fixture.workflowStep.Name,
			Status:     domain.WorkflowStepStatusFailed,
		})
		if err != nil {
			t.Fatalf("failed to create prior failed retry attempt: %v", err)
		}
	}

	failedWorkflow, err := fixture.workflowRepository.UpdateStatus(ctx, fixture.workflow.ID, domain.WorkflowStatusFailed)
	if err != nil {
		t.Fatalf("faile dto mark workflow failed: %v", err)
	}

	failedSteps, err := fixture.workflowStepRepository.ListByWorkflowID(ctx, fixture.workflow.ID)
	if err != nil {
		t.Fatalf("failed to load failed workflow steps: %v", err)
	}

	currentFailedStep := failedSteps[0]
	for _, candidate := range failedSteps {
		if candidate.ID == fixture.workflowStep.ID {
			currentFailedStep = candidate
			break
		}
	}

	_, err = executionService.RetryStep(ctx, failedWorkflow, currentFailedStep)
	if err == nil {
		t.Fatalf("expected retry policy to reject exceeded attempts")
	}
}
