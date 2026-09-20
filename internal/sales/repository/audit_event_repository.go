package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/platform/store"
)

type AuditEventRepository struct {
	store *store.PostgresStore
}

func NewAuditEventRepository(store *store.PostgresStore) *AuditEventRepository {
	return &AuditEventRepository{store: store}
}

func (r *AuditEventRepository) Create(ctx context.Context, event domain.AuditEvent) (domain.AuditEvent, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return domain.AuditEvent{}, fmt.Errorf("audit event repository is not configured")
	}

	const query = `
		INSERT INTO audit_events (id, workflow_id, event_type, actor)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`

	err := r.store.Pool().QueryRow(
		ctx,
		query,
		event.ID,
		event.WorkflowID,
		event.EventType,
		event.Actor,
	).Scan(&event.CreatedAt, &event.UpdatedAt)
	if err != nil {
		slog.Error("failed to create audit event", "error", err)
		return domain.AuditEvent{}, fmt.Errorf("failed to create audit event: %w", err)
	}

	return event, err
}

func (r *AuditEventRepository) ListByWorkflowID(ctx context.Context, workflowID string) ([]domain.AuditEvent, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return nil, fmt.Errorf("audit event repository is not configured")
	}

	const query = `
		SELECT id, workflow_id, event_type, actor, created_at, updated_at
		FROM audit_events
		WHERE workflow_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.store.Pool().Query(ctx, query, workflowID)
	if err != nil {
		slog.Error("failed to list audit events", "workflow_id", workflowID, "error", err)
		return nil, fmt.Errorf("failed to list audit events by workflow id: %w", err)
	}

	events := make([]domain.AuditEvent, 0)
	for rows.Next() {
		var event domain.AuditEvent
		if err := rows.Scan(&event.ID, &event.WorkflowID, &event.EventType, &event.Actor, &event.CreatedAt, &event.UpdatedAt); err != nil {
			slog.Error("failed to scan audit event", "error", err)
			return nil, fmt.Errorf("failed to scan audit event: %w", err)
		}
		events = append(events, event)
	}

	if rows.Err() != nil {
		slog.Error("failed to iterate audit events", "workflow_id", workflowID, "error", err)
		return nil, fmt.Errorf("failed to iterate audit events by workflow id: %w", err)
	}

	return events, nil
}
