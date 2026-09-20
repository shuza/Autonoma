package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/sales/repository"
)

type LeadService struct {
	companies *repository.CompanyRepository
	contacts  *repository.ContactRepository
	leads     *repository.LeadRepository
	workflows *repository.WorkflowRepository
}

func NewLeadService(companies *repository.CompanyRepository, contact *repository.ContactRepository, leads *repository.LeadRepository, workflows *repository.WorkflowRepository) *LeadService {
	return &LeadService{
		companies: companies,
		contacts:  contact,
		leads:     leads,
		workflows: workflows,
	}
}

func (s *LeadService) CreateLead(ctx context.Context, input CreateLeadInput) (CreateLeadResult, error) {
	if input.CompanyName == "" {
		return CreateLeadResult{}, fmt.Errorf("company name is required")
	}
	if input.ContactEmail == "" {
		return CreateLeadResult{}, fmt.Errorf("contact email is required")
	}

	company := domain.Company{
		ID:      uuid.NewString(),
		Name:    input.CompanyName,
		Website: input.Website,
	}

	createdCompany, err := s.companies.Create(ctx, company)
	if err != nil {
		return CreateLeadResult{}, fmt.Errorf("failed to create company: %w", err)
	}

	contact := domain.Contact{
		ID:        uuid.NewString(),
		CompanyID: createdCompany.ID,
		FirstName: input.ContactFirstName,
		LastName:  input.ContactLastName,
		Email:     input.ContactEmail,
	}

	createdContact, err := s.contacts.Create(ctx, contact)
	if err != nil {
		return CreateLeadResult{}, fmt.Errorf("failed to create contact: %w", err)
	}

	lead := domain.Lead{
		ID:        uuid.NewString(),
		CompanyID: createdCompany.ID,
		Source:    input.Source,
		Status:    domain.LeadStatusNew,
	}

	createdLead, err := s.leads.Create(ctx, lead)
	if err != nil {
		return CreateLeadResult{}, fmt.Errorf("failed to create lead: %w", err)
	}

	workflow := domain.Workflow{
		ID:     uuid.NewString(),
		LeadID: createdLead.ID,
		Status: domain.WorkflowStatusPending,
	}

	createdWorkflow, err := s.workflows.Create(ctx, workflow)
	if err != nil {
		return CreateLeadResult{}, fmt.Errorf("failed to create workflow: %w", err)
	}

	return CreateLeadResult{
		Company:  createdCompany,
		Contact:  createdContact,
		Lead:     createdLead,
		Workflow: createdWorkflow,
	}, nil
}

type CreateLeadInput struct {
	CompanyName      string
	Website          string
	Source           string
	ContactFirstName string
	ContactLastName  string
	ContactEmail     string
}

type CreateLeadResult struct {
	Company  domain.Company
	Contact  domain.Contact
	Lead     domain.Lead
	Workflow domain.Workflow
}
