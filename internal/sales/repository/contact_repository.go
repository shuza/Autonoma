package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/platform/store"
)

type ContactRepository struct {
	store *store.PostgresStore
}

func NewContactRepository(store *store.PostgresStore) *ContactRepository {
	return &ContactRepository{store: store}
}

func (r *ContactRepository) Create(ctx context.Context, contact domain.Contact) (domain.Contact, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return domain.Contact{}, fmt.Errorf("contact repository is not configured")
	}

	const query = `
		INSERT INTO contacts (id, company_id, first_name, last_name, email)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	err := r.store.Pool().QueryRow(
		ctx,
		query,
		contact.ID,
		contact.CompanyID,
		contact.FirstName,
		contact.LastName,
		contact.Email,
	).Scan(&contact.CreatedAt, &contact.UpdatedAt)

	if err != nil {
		slog.Error("failed to create contact", "error", err)
		return domain.Contact{}, fmt.Errorf("failed to create contact: %w", err)
	}

	return contact, nil
}

func (r *ContactRepository) GetByID(ctx context.Context, id string) (domain.Contact, error) {
	if r == nil || r.store == nil || r.store.Pool() == nil {
		return domain.Contact{}, fmt.Errorf("contact repository is not configured")
	}

	const query = `
		SELECT id, company_id, first_name, last_name, email, created_at, updated_at
		FROM contacts
		WHERE id = $1
	`

	var contact domain.Contact
	err := r.store.Pool().QueryRow(ctx, query, id).
		Scan(&contact.ID, &contact.CompanyID, &contact.FirstName, &contact.LastName, &contact.Email, &contact.CreatedAt, &contact.UpdatedAt)
	if err != nil {
		slog.Error("failed to get contact by id", "error", err)
		return domain.Contact{}, fmt.Errorf("failed to get contact by id: %w", err)
	}

	return contact, nil
}
