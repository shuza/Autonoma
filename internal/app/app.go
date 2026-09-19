package app

import (
	"context"
	"fmt"

	"github.com/shuza/Autonoma/internal/platform/config"
	"github.com/shuza/Autonoma/internal/platform/store"
	"github.com/shuza/Autonoma/internal/sales/repository"
	"github.com/shuza/Autonoma/internal/sales/service"
)

type App struct {
	LeadService *service.LeadService
	store       *store.PostgresStore
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	pgStore, err := store.NewPostgresStore(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres store: %w", err)
	}

	leadRepository := repository.NewLeadRepository(pgStore)
	workflowRepository := repository.NewWorkflowRepository(pgStore)
	companyRepository := repository.NewCompanyRepository(pgStore)
	contactRepository := repository.NewContactRepository(pgStore)

	return &App{
		LeadService: service.NewLeadService(companyRepository, contactRepository, leadRepository, workflowRepository),
		store:       pgStore,
	}, nil
}

func (a *App) Close() {
	if a != nil {
		a.store.Close()
	}
}
