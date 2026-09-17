package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/platform/store"
)

type LeadRepository struct {
	store *store.PostgresStore
}

func NewLeadRepository(store *store.PostgresStore) *LeadRepository {
	return &LeadRepository{store: store}
}

func (r *LeadRepository) Create(ctx context.Context, lead domain.Lead) (domain.Lead, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return domain.Lead{}, fmt.Errorf("lead repository is not configured")
	}

	const query = `
		INSERT INTO leads (id, company_name, website, source, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	err := r.store.Pool().QueryRow(
		ctx,
		query,
		lead.ID,
		lead.CompanyName,
		lead.Website,
		lead.Source,
		lead.Status,
	).Scan(&lead.CreatedAt, &lead.UpdatedAt)
	if err != nil {
		slog.Error("failed to create lead", "error", err)
		return domain.Lead{}, fmt.Errorf("failed to create lead: %w", err)
	}
	return lead, nil
}

func (r *LeadRepository) GetByID(ctx context.Context, id string) (domain.Lead, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return domain.Lead{}, fmt.Errorf("lead repository is not configured")
	}

	const query = `
		SELECT id, company_name, website, source, status, created_at, updated_at
		FROM leads
		WHERE id = $1
	`

	var lead domain.Lead
	err := r.store.Pool().QueryRow(
		ctx,
		query,
		id,
	).Scan(&lead.ID, &lead.CompanyName, &lead.Website, &lead.Source, &lead.Status, &lead.CreatedAt, &lead.UpdatedAt)
	if err != nil {
		slog.Error("failed to get lead by id", "error", err)
		return domain.Lead{}, fmt.Errorf("failed to get lead by id: %w", err)
	}
	return lead, nil
}
