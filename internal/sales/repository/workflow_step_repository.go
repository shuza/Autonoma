package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/platform/store"
)

type WorkflowStepRepository struct {
	store *store.PostgresStore
}

func NewWorkflowStepRepository(store *store.PostgresStore) *WorkflowStepRepository {
	return &WorkflowStepRepository{store: store}
}

func (r *WorkflowStepRepository) Create(ctx context.Context, step domain.WorkflowStep) (domain.WorkflowStep, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return domain.WorkflowStep{}, fmt.Errorf("workflow step repository is not configured")
	}

	const query = `
		INSERT INTO workflow_steps (id, workflow_id, name, status)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`

	err := r.store.Pool().QueryRow(
		ctx,
		query,
		step.ID,
		step.WorkflowID,
		step.Name,
		step.Status,
	).Scan(&step.CreatedAt, &step.UpdatedAt)
	if err != nil {
		slog.Error("failed to create workflow step", "error", err)
		return domain.WorkflowStep{}, fmt.Errorf("failed to create workflow step: %w", err)
	}

	return step, nil
}

func (r *WorkflowStepRepository) ListByWorkflowID(ctx context.Context, workflowID string) ([]domain.WorkflowStep, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return nil, fmt.Errorf("workflow step repository is not configured")
	}

	const query = `
		SELECT id, workflow_id, name, status, created_at, updated_at
		FROM workflow_steps
		WHERE workflow_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.store.Pool().Query(ctx, query, workflowID)
	if err != nil {
		slog.Error("failed to list workflow steps", "workflow_id", workflowID, "error", err)
		return nil, fmt.Errorf("failed to list workflow steps: %w", err)
	}
	defer rows.Close()

	steps := make([]domain.WorkflowStep, 0)
	for rows.Next() {
		var step domain.WorkflowStep
		if err := rows.Scan(&step.ID, &step.WorkflowID, &step.Name, &step.Status, &step.CreatedAt, &step.UpdatedAt); err != nil {
			slog.Error("failed to scan workflow step", "error", err)
			return nil, fmt.Errorf("failed to scan workflow step: %w", err)
		}
		steps = append(steps, step)
	}

	if err := rows.Err(); err != nil {
		slog.Error("failed to iterate workflow steps", "workflow_id", workflowID, "error", err)
		return nil, fmt.Errorf("failed to iterate workflow steps by workflow id: %w", err)
	}

	return steps, nil
}

func (r *WorkflowStepRepository) UpdateStatus(ctx context.Context, stepID string, status domain.WorkflowStepStatus) (domain.WorkflowStep, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return domain.WorkflowStep{}, fmt.Errorf("workflow step repository is not configured")
	}

	const query = `
		UPDATE workflow_steps
		SET status = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, workflow_id, name, status, created_at, updated_at
	`

	var step domain.WorkflowStep
	err := r.store.Pool().QueryRow(ctx, query, stepID, status).
		Scan(&step.ID, &step.WorkflowID, &step.Name, &step.Status, &step.CreatedAt, &step.UpdatedAt)
	if err != nil {
		slog.Error("failed to update workflow step status", "step_id", stepID, "status", status, "error", err)
		return domain.WorkflowStep{}, fmt.Errorf("failed to update workflow step status: %w", err)
	}

	return step, nil
}
