package integration

import (
	"testing"

	"github.com/shuza/Autonoma/internal/sales/repository"
)

func TestLeadWorkflowBusinessContextIsPersisted(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)
	pgStore := newIntegrationStore(t, ctx, databaseURL)
	defer pgStore.Close()

	fixture := createPersistenceFixture(t, ctx, pgStore)

	storedLead, err := fixture.leadRepository.GetByID(ctx, fixture.lead.ID)
	if err != nil {
		t.Fatalf("failed to get lead by id: %v", err)
	}

	if storedLead.CompanyID != fixture.lead.CompanyID {
		t.Fatalf("expected company name %q, got %q", fixture.lead.CompanyID, storedLead.CompanyID)
	}

	if storedLead.CreatedAt.IsZero() || storedLead.UpdatedAt.IsZero() {
		t.Fatal("expected postgres-managed timestamps to be populated for lead")
	}

	storedCompany, err := fixture.companyRepository.GetByID(ctx, fixture.company.ID)
	if err != nil {
		t.Fatalf("failed to get company by id: %v", err)
	}

	if storedCompany.Name != fixture.company.Name {
		t.Fatalf("expected company name %q, got %q", fixture.company.Name, storedCompany.Name)
	}

	if storedCompany.CreatedAt.IsZero() || storedCompany.UpdatedAt.IsZero() {
		t.Fatal("expected postgres-managed timestamps to be populated for company")
	}

	workflows, err := fixture.workflowRepository.ListByLeadID(ctx, fixture.lead.ID)
	if err != nil {
		t.Fatalf("failed to list workflows by lead id: %v", err)
	}

	if len(workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(workflows))
	}

	if workflows[0].Status != fixture.workflow.Status {
		t.Fatalf("expected workflow status %q, got %q", fixture.workflow.Status, workflows[0].Status)
	}

	if workflows[0].CreatedAt.IsZero() || workflows[0].UpdatedAt.IsZero() {
		t.Fatal("expected postgres-managed timestamps to be populated for workflow")
	}

	steps, err := fixture.workflowStepRepository.ListByWorkflowID(ctx, fixture.workflow.ID)
	if err != nil {
		t.Fatalf("failed to list workflow steps by workflow id: %v", err)
	}

	if len(steps) != 1 {
		t.Fatalf("expected 1 workflow step, got %d", len(steps))
	}

	if steps[0].ID != fixture.workflowStep.ID || steps[0].Name != fixture.workflowStep.Name || steps[0].Status != fixture.workflowStep.Status {
		t.Fatalf("unexpected workflow step payload: %+v", steps[0])
	}

	if steps[0].CreatedAt.IsZero() || steps[0].UpdatedAt.IsZero() {
		t.Fatal("expected postgres-managed timestamps to be populated for workflow step")
	}

	toolExecutions, err := fixture.toolExecutionRepository.ListByWorkflowStepID(ctx, fixture.workflowStep.ID)
	if err != nil {
		t.Fatalf("failed to list tool executions by workflow step id: %v", err)
	}

	if len(toolExecutions) != 1 {
		t.Fatalf("expected 1 tool execution, got %d", len(toolExecutions))
	}

	if toolExecutions[0].ID != fixture.toolExecution.ID || toolExecutions[0].WorkflowStepID != fixture.toolExecution.WorkflowStepID || toolExecutions[0].ToolName != fixture.toolExecution.ToolName || toolExecutions[0].Status != fixture.toolExecution.Status {
		t.Fatalf("unexpected tool execution payload: %+v", toolExecutions[0])
	}

	if toolExecutions[0].CreatedAt.IsZero() || toolExecutions[0].UpdatedAt.IsZero() {
		t.Fatal("expected postgres-managed timestamps to be populated for tool execution")
	}

	approvals, err := fixture.approvalRepository.ListByToolExecutionID(ctx, fixture.toolExecution.ID)
	if err != nil {
		t.Fatalf("failed to list approvals by tool execution id: %v", err)
	}

	if len(approvals) != 1 {
		t.Fatalf("expected 1 approval, got %d", len(approvals))
	}

	if approvals[0].ID != fixture.approval.ID || approvals[0].ToolExecutionID != fixture.approval.ToolExecutionID || approvals[0].RequestedBy != fixture.approval.RequestedBy || approvals[0].Status != fixture.approval.Status {
		t.Fatalf("unexpected approval payload: %+v", approvals[0])
	}

	if approvals[0].CreatedAt.IsZero() || approvals[0].UpdatedAt.IsZero() {
		t.Fatal("expected postgres-managed timestamps to be populated for approval")
	}

	auditEvents, err := fixture.auditEventRepository.ListByWorkflowID(ctx, fixture.workflow.ID)
	if err != nil {
		t.Fatalf("failed to list audit events by workflow id: %v", err)
	}

	if len(auditEvents) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditEvents))
	}

	if auditEvents[0].ID != fixture.auditEvent.ID || auditEvents[0].WorkflowID != fixture.auditEvent.WorkflowID || auditEvents[0].EventType != fixture.auditEvent.EventType || auditEvents[0].Actor != fixture.auditEvent.Actor {
		t.Fatalf("unexpected audit event payload: %+v", auditEvents[0])
	}

	if auditEvents[0].CreatedAt.IsZero() || auditEvents[0].UpdatedAt.IsZero() {
		t.Fatal("expected postgres-managed timestamps to be populated for audit event")
	}
}

func TestLeadWorkflowBusinessContextSurvivesProcessRestart(t *testing.T) {
	ctx, databaseURL := setupIntegrationDatabase(t)

	firstStore := newIntegrationStore(t, ctx, databaseURL)
	fixture := createPersistenceFixture(t, ctx, firstStore)
	firstStore.Close()

	restartedStore := newIntegrationStore(t, ctx, databaseURL)
	defer restartedStore.Close()

	leadRepository := repository.NewLeadRepository(restartedStore)
	workflowRepository := repository.NewWorkflowRepository(restartedStore)
	workflowStepRepository := repository.NewWorkflowStepRepository(restartedStore)
	toolExecutionRepository := repository.NewToolExecutionRepository(restartedStore)
	approvalRepository := repository.NewApprovalRepository(restartedStore)
	auditEventRepository := repository.NewAuditEventRepository(restartedStore)
	companyRepository := repository.NewCompanyRepository(restartedStore)

	company, err := companyRepository.GetByID(ctx, fixture.company.ID)
	if err != nil {
		t.Fatalf("failed to reload company after restart: %v", err)
	}
	if company.Name != fixture.company.Name || company.Website != fixture.company.Website {
		t.Fatalf("unexpected company after restart: %+v", err)
	}

	lead, err := leadRepository.GetByID(ctx, fixture.lead.ID)
	if err != nil {
		t.Fatalf("failed to reload lead after restart: %v", err)
	}
	if lead.CompanyID != fixture.lead.CompanyID || lead.Source != fixture.lead.Source || lead.Status != fixture.lead.Status {
		t.Fatalf("unexpected lead after restart: %+v", err)
	}

	workflows, err := workflowRepository.ListByLeadID(ctx, fixture.lead.ID)
	if err != nil {
		t.Fatalf("failed to reload workflow after restart: %v", err)
	}
	if len(workflows) != 1 || workflows[0].ID != fixture.workflow.ID || workflows[0].Status != fixture.workflow.Status {
		t.Fatalf("unexpected workflow after restart: %+v", err)
	}

	steps, err := workflowStepRepository.ListByWorkflowID(ctx, fixture.workflow.ID)
	if err != nil {
		t.Fatalf("failed to reload workflow steps after restart: %v", err)
	}
	if len(steps) != 1 || steps[0].ID != fixture.workflowStep.ID || steps[0].Status != fixture.workflowStep.Status {
		t.Fatalf("unexpected workflow step after restart: %+v", err)
	}

	toolExecution, err := toolExecutionRepository.ListByWorkflowStepID(ctx, fixture.workflowStep.ID)
	if err != nil {
		t.Fatalf("failed to reload tool execution after restart: %v", err)
	}
	if len(toolExecution) != 1 || toolExecution[0].ID != fixture.toolExecution.ID || toolExecution[0].Status != fixture.toolExecution.Status {
		t.Fatalf("unexpected tool execution after restart: %+v", err)
	}

	approvals, err := approvalRepository.ListByToolExecutionID(ctx, fixture.toolExecution.ID)
	if err != nil {
		t.Fatalf("failed to reload approvals after restart: %+v", err)
	}
	if len(approvals) != 1 || approvals[0].ID != fixture.approval.ID || approvals[0].Status != fixture.approval.Status {
		t.Fatalf("unexpected approvals after restart: %+v", err)
	}

	auditEvent, err := auditEventRepository.ListByWorkflowID(ctx, fixture.workflow.ID)
	if err != nil {
		t.Fatalf("failed to reload audit event after restart: %+v", err)
	}
	if len(auditEvent) != 1 || auditEvent[0].ID != fixture.auditEvent.ID || auditEvent[0].EventType != fixture.auditEvent.EventType {
		t.Fatalf("unexpected audit event after restart: %+v", err)
	}
}
