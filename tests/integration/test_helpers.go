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

func (p persistenceFixture) newCompanyRepository(pgStore *store.PostgresStore) *repository.CompanyRepository {
	return repository.NewCompanyRepository(pgStore)
}

func (p persistenceFixture) newLeadRepository(pgStore *store.PostgresStore) *repository.LeadRepository {
	return repository.NewLeadRepository(pgStore)
}

func (p persistenceFixture) newWorkflowRepository(pgStore *store.PostgresStore) *repository.WorkflowRepository {
	return repository.NewWorkflowRepository(pgStore)
}

func (p persistenceFixture) newWorkflowStepRepository(pgStore *store.PostgresStore) *repository.WorkflowStepRepository {
	return repository.NewWorkflowStepRepository(pgStore)
}

func (p persistenceFixture) newToolExecutionRepository(pgStore *store.PostgresStore) *repository.ToolExecutionRepository {
	return repository.NewToolExecutionRepository(pgStore)
}

func (p persistenceFixture) newApprovalRepository(pgStore *store.PostgresStore) *repository.ApprovalRepository {
	return repository.NewApprovalRepository(pgStore)
}

func (p persistenceFixture) newAuditEventRepository(pgStore *store.PostgresStore) *repository.AuditEventRepository {
	return repository.NewAuditEventRepository(pgStore)
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
