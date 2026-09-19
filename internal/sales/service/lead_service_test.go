package service

import (
	"context"
	"testing"

	"github.com/shuza/Autonoma/internal/sales/repository"
)

func TestCreateLeadRequiresCompanyName(t *testing.T) {
	t.Parallel()

	service := NewLeadService(&repository.CompanyRepository{}, &repository.ContactRepository{}, &repository.LeadRepository{}, &repository.WorkflowRepository{})
	_, err := service.CreateLead(context.Background(), CreateLeadInput{})
	if err == nil {
		t.Fatalf("expected error for missing company name")
	}
}

func TestCreateLeadRequiresContactEmails(t *testing.T) {
	t.Parallel()

	service := NewLeadService(&repository.CompanyRepository{}, &repository.ContactRepository{}, &repository.LeadRepository{}, &repository.WorkflowRepository{})
	_, err := service.CreateLead(context.Background(), CreateLeadInput{
		CompanyName:      "Acme",
		Website:          "http://acme.test",
		Source:           "manual",
		ContactFirstName: "Ada",
		ContactLastName:  "Lovelace",
	})

	if err == nil {
		t.Fatalf("expected error for missing contact emails")
	}
}

func TestCreateLeadReturnsRepositoryErrors(t *testing.T) {
	t.Parallel()

	service := NewLeadService(&repository.CompanyRepository{}, &repository.ContactRepository{}, &repository.LeadRepository{}, &repository.WorkflowRepository{})
	_, err := service.CreateLead(context.Background(), CreateLeadInput{
		CompanyName:      "Acme",
		Website:          "http://acme.test",
		Source:           "manual",
		ContactFirstName: "Ada",
		ContactLastName:  "Lovelace",
		ContactEmail:     "ada@acme.test",
	})
	if err == nil {
		t.Fatalf("expected repository configuration error")
	}
}
