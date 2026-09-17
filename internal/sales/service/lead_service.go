package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/sales/repository"
)

type LeadService struct {
	leads     *repository.LeadRepository
	workflows *repository.WorkflowRepository
}

func NewLeadService(leads *repository.LeadRepository, workflows *repository.WorkflowRepository) *LeadService {
	return &LeadService{
		leads:     leads,
		workflows: workflows,
	}
}

func (s *LeadService) CreateLead(ctx context.Context, input CreateLeadInput) (CreateLeadResult, error) {
	if input.CompanyName == "" {
		return CreateLeadResult{}, fmt.Errorf("company name is required")
	}

	lead := domain.Lead{
		ID:          uuid.NewString(),
		CompanyName: input.CompanyName,
		Website:     input.Website,
		Source:      input.Source,
		Status:      domain.LeadStatusNew,
	}

	createdLead, err := s.leads.Create(ctx, lead)
	if err != nil {
		return CreateLeadResult{}, fmt.Errorf("failed to create lead: %w", err)
	}

	workflow := domain.Workflow{
		ID:     uuid.NewString(),
		LeadID: createdLead.ID,
		Status: domain.WorkflowStatusNew,
	}

	createdWorkflow, err := s.workflows.Create(ctx, workflow)
	if err != nil {
		return CreateLeadResult{}, fmt.Errorf("failed to create workflow: %w", err)
	}

	return CreateLeadResult{
		Lead:     createdLead,
		Workflow: createdWorkflow,
	}, nil
}

type CreateLeadInput struct {
	CompanyName string
	Website     string
	Source      string
}

type CreateLeadResult struct {
	Lead     domain.Lead
	Workflow domain.Workflow
}
