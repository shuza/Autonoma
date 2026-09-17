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

	lead := domain.Lead{
		ID:          uuid.NewString(),
		CompanyName: "Acme Corp",
		Website:     "https://acme.test",
		Source:      "manual",
		Status:      domain.LeadStatusNew,
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

	if _, err := workflowRepository.Create(ctx, workflow); err != nil {
		t.Fatalf("failed to create workflow: %v", err)
	}

	storedLead, err := leadRepository.GetByID(ctx, createdLead.ID)
	if err != nil {
		t.Fatalf("failed to get lead by id: %v", err)
	}

	if storedLead.CompanyName != lead.CompanyName {
		t.Fatalf("expected company name %q, got %q", lead.CompanyName, storedLead.CompanyName)
	}

	if storedLead.CreatedAt.IsZero() || storedLead.UpdatedAt.IsZero() {
		t.Fatal("expected postgres-managed timestamps to be populated for lead")
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
