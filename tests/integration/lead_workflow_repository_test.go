package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/platform/store"
	"github.com/shuza/Autonoma/internal/sales/repository"
)

func TestLeadAndWorkflowRepositoriesPersistRecords(t *testing.T) {
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

func TestLeadWorkflowStateAndBusinessContextSurviveProcessRestart(t *testing.T) {
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

type persistenceFixture struct {
	company                 domain.Company
	lead                    domain.Lead
	workflow                domain.Workflow
	workflowStep            domain.WorkflowStep
	toolExecution           domain.ToolExecution
	approval                domain.Approval
	auditEvent              domain.AuditEvent
	companyRepository       *repository.CompanyRepository
	leadRepository          *repository.LeadRepository
	workflowRepository      *repository.WorkflowRepository
	workflowStepRepository  *repository.WorkflowStepRepository
	toolExecutionRepository *repository.ToolExecutionRepository
	approvalRepository      *repository.ApprovalRepository
	auditEventRepository    *repository.AuditEventRepository
}

func setupIntegrationDatabase(t *testing.T) (context.Context, string) {
	t.Helper()

	databaseURL := os.Getenv("AUTONOMA_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("AUTONOMA_DATABASE_URL is required for integration tests")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	migrationsPath := filepath.Join("..", "..", "migrations")
	applyMigrations(t, ctx, conn, migrationsPath, ".down.sql")
	applyMigrations(t, ctx, conn, migrationsPath, ".up.sql")
	t.Cleanup(func() {
		applyMigrations(t, ctx, conn, migrationsPath, ".down.sql")
		conn.Close(ctx)
	})

	return ctx, databaseURL
}

func newIntegrationStore(t *testing.T, ctx context.Context, databaseURL string) *store.PostgresStore {
	t.Helper()

	pgStore, err := store.NewPostgresStore(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to create postgres store: %v", err)
	}

	return pgStore
}

func createPersistenceFixture(t *testing.T, ctx context.Context, pgStore *store.PostgresStore) persistenceFixture {
	t.Helper()

	leadRepository := repository.NewLeadRepository(pgStore)
	workflowRepository := repository.NewWorkflowRepository(pgStore)
	workflowStepRepository := repository.NewWorkflowStepRepository(pgStore)
	toolExecutionRepository := repository.NewToolExecutionRepository(pgStore)
	approvalRepository := repository.NewApprovalRepository(pgStore)
	auditEventRepository := repository.NewAuditEventRepository(pgStore)
	companyRepository := repository.NewCompanyRepository(pgStore)

	company := domain.Company{
		ID:      uuid.NewString(),
		Name:    "Acme Corp",
		Website: "https://acme.test",
	}

	createdCompany, err := companyRepository.Create(ctx, company)
	if err != nil {
		t.Fatalf("failed to create company: %v", err)
	}

	lead := domain.Lead{
		ID:        uuid.NewString(),
		CompanyID: createdCompany.ID,
		Source:    "manual",
		Status:    domain.LeadStatusNew,
	}

	createdLead, err := leadRepository.Create(ctx, lead)
	if err != nil {
		t.Fatalf("failed to create lead: %v", err)
	}

	workflow := domain.Workflow{
		ID:     uuid.NewString(),
		LeadID: createdLead.ID,
		Status: domain.WorkflowStatusPending,
	}

	createdWorkflow, err := workflowRepository.Create(ctx, workflow)
	if err != nil {
		t.Fatalf("failed to create workflow: %v", err)
	}

	step := domain.WorkflowStep{
		ID:         uuid.NewString(),
		WorkflowID: createdWorkflow.ID,
		Name:       "load-context",
		Status:     domain.WorkflowStepStatusPending,
	}

	createdStep, err := workflowStepRepository.Create(ctx, step)
	if err != nil {
		t.Fatalf("failed to create workflow step: %v", err)
	}

	toolExecution := domain.ToolExecution{
		ID:             uuid.NewString(),
		WorkflowStepID: createdStep.ID,
		ToolName:       "web_search",
		Status:         domain.ToolExecutionStatusPending,
	}

	createdToolExecution, err := toolExecutionRepository.Create(ctx, toolExecution)
	if err != nil {
		t.Fatalf("failed to create tool execution: %v", err)
	}

	approval := domain.Approval{
		ID:              uuid.NewString(),
		ToolExecutionID: createdToolExecution.ID,
		Status:          domain.ApprovalStatusPending,
		RequestedBy:     "policy-engine",
	}

	createdApproval, err := approvalRepository.Create(ctx, approval)
	if err != nil {
		t.Fatalf("failed to create approval: %v", err)
	}

	auditEvent := domain.AuditEvent{
		ID:         uuid.NewString(),
		WorkflowID: createdWorkflow.ID,
		EventType:  "approval_requested",
		Actor:      "policy-engine",
	}

	createdAuditEvent, err := auditEventRepository.Create(ctx, auditEvent)
	if err != nil {
		t.Fatalf("failed to create audit event: %v", err)
	}

	return persistenceFixture{
		company:                 createdCompany,
		lead:                    createdLead,
		workflow:                createdWorkflow,
		workflowStep:            createdStep,
		toolExecution:           createdToolExecution,
		approval:                createdApproval,
		auditEvent:              createdAuditEvent,
		companyRepository:       companyRepository,
		leadRepository:          leadRepository,
		workflowRepository:      workflowRepository,
		workflowStepRepository:  workflowStepRepository,
		toolExecutionRepository: toolExecutionRepository,
		approvalRepository:      approvalRepository,
		auditEventRepository:    auditEventRepository,
	}
}

func applyMigrations(t *testing.T, ctx context.Context, conn *pgx.Conn, dirPath string, suffix string) {
	t.Helper()

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		t.Fatalf("failed to read migrations dir %s: %v", dirPath, err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), suffix) {
			continue
		}

		paths = append(paths, filepath.Join(dirPath, entry.Name()))
	}

	sort.Strings(paths)
	if suffix == ".down.sql" {
		sort.Sort(sort.Reverse(sort.StringSlice(paths)))
	}

	if len(paths) == 0 {
		t.Fatalf("no migration files matching %q found in %s", suffix, dirPath)
	}

	for _, path := range paths {
		applyMigrationFile(t, ctx, conn, path)
	}
}

func applyMigrationFile(t *testing.T, ctx context.Context, conn *pgx.Conn, path string) {
	t.Helper()

	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read migration file: %v", err)
	}

	if _, err := conn.Exec(ctx, string(sqlBytes)); err != nil {
		t.Fatalf("apply migration %s: %v", path, fmt.Errorf("failed exec migration: %w", err))
	}
}
