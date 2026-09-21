package domain

import "testing"

func TestWorkflowTransitionToAllowsValidTransitions(t *testing.T) {
	t.Parallel()

	workflow := &Workflow{Status: WorkflowStatusPending}
	if err := workflow.TransitionTo(WorkflowStatusRunning); err != nil {
		t.Fatalf("expected pending to running transition to succeed: %v", err)
	}

	if workflow.Status != WorkflowStatusRunning {
		t.Fatalf("expected workflow status to be %s, got %s", WorkflowStatusRunning, workflow.Status)
	}
}

func TestWorkflowTransitionToRejectsInvalidTransitions(t *testing.T) {
	t.Parallel()
	workflow := &Workflow{Status: WorkflowStatusPending}
	if err := workflow.TransitionTo(WorkflowStatusCompleted); err == nil {
		t.Fatalf("expected pending to completed transition to fail")
	}
}

func TestWorkflowTransitionToAllowsWaitingApprovalCycle(t *testing.T) {
	t.Parallel()

	workflow := &Workflow{Status: WorkflowStatusRunning}
	if err := workflow.TransitionTo(WorkflowStatusWaitingApproval); err != nil {
		t.Fatalf("expected running to waiting approval transition to succeed: %v", err)
	}

	if err := workflow.TransitionTo(WorkflowStatusRunning); err != nil {
		t.Fatalf("expected waiting approval to running transition to succeed: %v", err)
	}
}
