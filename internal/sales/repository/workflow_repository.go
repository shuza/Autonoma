package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/platform/store"
)

type WorkflowRepository struct {
	store *store.PostgresStore
}

func NewWorkflowRepository(store *store.PostgresStore) *WorkflowRepository {
	return &WorkflowRepository{store: store}
}

func (r *WorkflowRepository) Create(ctx context.Context, workflow domain.Workflow) (domain.Workflow, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return domain.Workflow{}, errors.New("workflow repository is not configured")
	}

	const query = `
		INSERT INTO workflows (id, lead_id, status)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at
	`

	err := r.store.Pool().QueryRow(
		ctx,
		query,
		workflow.ID,
		workflow.LeadID,
		workflow.Status,
	).Scan(&workflow.CreatedAt, &workflow.UpdatedAt)
	if err != nil {
		slog.Error("failed to create workflow", "error", err)
		return domain.Workflow{}, fmt.Errorf("failed to create workflow: %w", err)
	}

	return workflow, nil
}

func (r *WorkflowRepository) ListByLeadID(ctx context.Context, leadID string) ([]domain.Workflow, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return nil, errors.New("workflow repository is not configured")
	}

	const query = `
		SELECT id, lead_id, status, created_at, updated_at
		FROM workflows
		WHERE lead_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.store.Pool().Query(ctx, query, leadID)
	if err != nil {
		slog.Error("failed to list workflows", "lead_id", leadID, "error", err)
		return nil, fmt.Errorf("failed to list workflow by lead id: %w", err)
	}
	defer rows.Close()

	workflows := make([]domain.Workflow, 0)
	for rows.Next() {
		var workflow domain.Workflow
		if err := rows.Scan(&workflow.ID, &workflow.LeadID, &workflow.Status, &workflow.CreatedAt, &workflow.UpdatedAt); err != nil {
			slog.Error("failed to scan workflow", "error", err)
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}
		workflows = append(workflows, workflow)
	}

	if err := rows.Err(); err != nil {
		slog.Error("failed to list workflows", "lead_id", leadID, "error", err)
		return nil, fmt.Errorf("failed to iterate workflow by lead id: %w", err)
	}

	return workflows, nil
}

func (r *WorkflowRepository) UpdateStatus(ctx context.Context, workflowID string, status domain.WorkflowStatus) (domain.Workflow, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return domain.Workflow{}, errors.New("workflow repository is not configured")
	}

	const query = `
		Update workflows
		SET status = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, lead_id, status, created_at, updated_at
	`

	var workflow domain.Workflow
	err := r.store.Pool().QueryRow(ctx, query, workflowID, status).
		Scan(&workflow.ID, &workflow.LeadID, &workflow.Status, &workflow.CreatedAt, &workflow.UpdatedAt)
	if err != nil {
		slog.Error("failed to update workflow status", "workflow_id", workflowID, "status", status, "error", err)
		return domain.Workflow{}, fmt.Errorf("failed to update workflow status: %w", err)
	}

	return workflow, nil
}
