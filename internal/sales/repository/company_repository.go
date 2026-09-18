package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/platform/store"
)

type CompanyRepository struct {
	store *store.PostgresStore
}

func NewCompanyRepository(store *store.PostgresStore) *CompanyRepository {
	return &CompanyRepository{
		store: store,
	}
}

func (r *CompanyRepository) Create(ctx context.Context, company domain.Company) (domain.Company, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return domain.Company{}, fmt.Errorf("company repository is not configured")
	}

	const query = `
		INSERT INTO companies (id, name, website)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at
	`

	err := r.store.Pool().QueryRow(
		ctx,
		query,
		company.ID,
		company.Name,
		company.Website,
	).Scan(&company.CreatedAt, &company.UpdatedAt)

	if err != nil {
		slog.Error("failed to create company", "error", err)
		return domain.Company{}, fmt.Errorf("failed to create company: %w", err)
	}

	return company, nil
}

func (r *CompanyRepository) GetByID(ctx context.Context, id string) (domain.Company, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return domain.Company{}, fmt.Errorf("company repository is not configured")
	}

	const query = `
		SELECT id, name, website, created_at, updated_at
		FROM companies
		WHERE id = $1
	`

	var company domain.Company
	err := r.store.Pool().QueryRow(ctx, query, id).Scan(&company.ID, &company.Name, &company.Website, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		slog.Error("failed to get company by id", "error", err)
		return domain.Company{}, fmt.Errorf("failed to get company by id: %w", err)
	}

	return company, nil
}
