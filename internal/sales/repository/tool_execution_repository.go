package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/platform/store"
)

type ToolExecutionRepository struct {
	store *store.PostgresStore
}

func NewToolExecutionRepository(store *store.PostgresStore) *ToolExecutionRepository {
	return &ToolExecutionRepository{store: store}
}

func (r *ToolExecutionRepository) Create(ctx context.Context, execution domain.ToolExecution) (domain.ToolExecution, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return domain.ToolExecution{}, fmt.Errorf("tool execution repository is not configured")
	}

	const query = `
		INSERT INTO tool_executions (id, workflow_step_id, tool_name, status)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`

	err := r.store.Pool().QueryRow(
		ctx,
		query,
		execution.ID,
		execution.WorkflowStepID,
		execution.ToolName,
		execution.Status,
	).Scan(&execution.CreatedAt, execution.UpdatedAt)
	if err != nil {
		slog.Error("failed to create tool execution", "error", err)
		return domain.ToolExecution{}, fmt.Errorf("failed to create tool execution: %w", err)
	}

	return execution, nil
}

func (r *ToolExecutionRepository) ListByWorkflowStepID(ctx context.Context, workflowStepID string) ([]domain.ToolExecution, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return nil, fmt.Errorf("tool execution repository is not configured")
	}

	const query = `
		SELECT id, workflow_step_id, tool_name, status, created_at, updated_at
		FROM tool_executions
		WHERE workflow_step_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.store.Pool().Query(ctx, query, workflowStepID)
	if err != nil {
		slog.Error("failed to list tool executions", "workflow_step_id", workflowStepID, "error", err)
		return nil, fmt.Errorf("failed to list tool executions by workflow step id: %w", err)
	}
	defer rows.Close()

	executions := make([]domain.ToolExecution, 0)
	for rows.Next() {
		var execution domain.ToolExecution
		if err := rows.Scan(&execution.ID, &execution.WorkflowStepID, &execution.ToolName, &execution.Status, &execution.CreatedAt, &execution.UpdatedAt); err != nil {
			slog.Error("failed to scan tool execution", "error", err)
			return nil, fmt.Errorf("failed to scan tool execution: %w", err)
		}
		executions = append(executions, execution)
	}

	if err := rows.Err(); err != nil {
		slog.Error("failed to list tool executions", "workflow_step_id", workflowStepID, "error", err)
		return nil, fmt.Errorf("failed to list tool executions by workflow step id: %w", err)
	}

	return executions, nil
}
