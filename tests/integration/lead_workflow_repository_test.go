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
	databaseURL := os.Getenv("AUTONOMA_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("AUTONOMA_DATABASE_URL is required for integration tests")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to connect postgres: %v", err)
	}
	migrationsPath := filepath.Join("..", "..", "migrations")
	applyMigrations(t, ctx, conn, migrationsPath, ".down.sql")
	applyMigrations(t, ctx, conn, migrationsPath, ".up.sql")
	t.Cleanup(func() {
		applyMigrations(t, ctx, conn, migrationsPath, ".down.sql")
		conn.Close(ctx)
	})

	pgStore, err := store.NewPostgresStore(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to create postgres store: %v", err)
	}
	defer pgStore.Close()

	leadRepository := repository.NewLeadRepository(pgStore)
	workflowRepository := repository.NewWorkflowRepository(pgStore)
	workflowStepRepository := repository.NewWorkflowStepRepository(pgStore)
	toolExecutionRepository := repository.NewToolExecutionRepository(pgStore)
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
		Status: domain.WorkflowStatusNew,
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

	storedLead, err := leadRepository.GetByID(ctx, createdLead.ID)
	if err != nil {
		t.Fatalf("failed to get lead by id: %v", err)
	}

	if storedLead.CompanyID != lead.CompanyID {
		t.Fatalf("expected company name %q, got %q", lead.CompanyID, storedLead.CompanyID)
	}

	if storedLead.CreatedAt.IsZero() || storedLead.UpdatedAt.IsZero() {
		t.Fatal("expected postgres-managed timestamps to be populated for lead")
	}

	storedCompany, err := companyRepository.GetByID(ctx, createdCompany.ID)
	if err != nil {
		t.Fatalf("failed to get company by id: %v", err)
	}

	if storedCompany.Name != createdCompany.Name {
		t.Fatalf("expected company name %q, got %q", createdCompany.Name, storedCompany.Name)
	}

	if storedCompany.CreatedAt.IsZero() || storedCompany.UpdatedAt.IsZero() {
		t.Fatal("expected postgres-managed timestamps to be populated for company")
	}

	workflows, err := workflowRepository.ListByLeadID(ctx, createdLead.ID)
	if err != nil {
		t.Fatalf("failed to list workflows by lead id: %v", err)
	}

	if len(workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(workflows))
	}

	if workflows[0].CreatedAt.IsZero() || workflows[0].UpdatedAt.IsZero() {
		t.Fatal("expected postgres-managed timestamps to be populated for workflow")
	}

	steps, err := workflowStepRepository.ListByWorkflowID(context.Background(), createdWorkflow.ID)
	if err != nil {
		t.Fatalf("failed to list workflow steps by workflow id: %v", err)
	}

	if len(steps) != 1 {
		t.Fatalf("expected 1 workflow step, got %d", len(steps))
	}

	if steps[0].ID != createdStep.ID || steps[0].Name != createdStep.Name || steps[0].Status != createdStep.Status {
		t.Fatalf("unexpected workflow step payload: %+v", steps[0])
	}

	if steps[0].CreatedAt.IsZero() || steps[0].UpdatedAt.IsZero() {
		t.Fatal("expected postgres-managed timestamps to be populated for workflow step")
	}

	toolExecutions, err := toolExecutionRepository.ListByWorkflowStepID(ctx, createdStep.ID)
	if err != nil {
		t.Fatalf("failed to list tool executions by workflow step id: %v", err)
	}

	if len(toolExecutions) != 1 {
		t.Fatalf("expected 1 tool execution, got %d", len(toolExecutions))
	}

	if toolExecutions[0].ID != createdToolExecution.ID || toolExecutions[0].WorkflowStepID != createdToolExecution.WorkflowStepID || toolExecutions[0].ToolName != createdToolExecution.ToolName || toolExecutions[0].Status != createdToolExecution.Status {
		t.Fatalf("unexpected tool execution payload: %+v", toolExecutions[0])
	}

	if toolExecutions[0].CreatedAt.IsZero() || toolExecutions[0].UpdatedAt.IsZero() {
		t.Fatal("expected postgres-managed timestamps to be populated for tool execution")
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
