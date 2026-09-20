package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/platform/store"
)

type ApprovalRepository struct {
	store *store.PostgresStore
}

func NewApprovalRepository(store *store.PostgresStore) *ApprovalRepository {
	return &ApprovalRepository{store: store}
}

func (r *ApprovalRepository) Create(ctx context.Context, approval domain.Approval) (domain.Approval, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return domain.Approval{}, fmt.Errorf("approval repository is not configured")
	}

	const query = `
		INSERT INTO approvals (id, tool_execution_id, status, requested_by)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`

	err := r.store.Pool().QueryRow(
		ctx,
		query,
		approval.ID,
		approval.ToolExecutionID,
		approval.Status,
		approval.RequestedBy,
	).Scan(&approval.CreatedAt, &approval.UpdatedAt)
	if err != nil {
		slog.Error("failed to create approval", "error", err)
		return domain.Approval{}, fmt.Errorf("failed to create approval: %w", err)
	}

	return approval, nil
}

func (r *ApprovalRepository) ListByToolExecutionID(ctx context.Context, toolExecutionID string) ([]domain.Approval, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return nil, fmt.Errorf("approval repository is not configured")
	}

	const query = `
		SELECT id, tool_execution_id, status, requested_by, created_at, updated_at
		FROM approvals
		WHERE tool_execution_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.store.Pool().Query(ctx, query, toolExecutionID)
	if err != nil {
		slog.Error("failed to list approvals", "tool_execution_id", toolExecutionID, "error", err)
		return nil, fmt.Errorf("failed to list approvals by tool execution id: %w", err)
	}
	defer rows.Close()

	approvals := make([]domain.Approval, 0)
	for rows.Next() {
		var approval domain.Approval
		if err := rows.Scan(&approval.ID, &approval.ToolExecutionID, &approval.Status, &approval.RequestedBy, &approval.CreatedAt, &approval.UpdatedAt); err != nil {
			slog.Error("failed to scan approval", "error", err)
			return nil, fmt.Errorf("failed to scan approval: %w", err)
		}
		approvals = append(approvals, approval)
	}

	if err := rows.Err(); err != nil {
		slog.Error("failed to iterate approvals", "tool_execution_id", toolExecutionID, "error", err)
		return nil, fmt.Errorf("failed to iterate approvals: %w", err)
	}

	return approvals, nil
}
