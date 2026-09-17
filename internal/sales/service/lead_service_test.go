package service

import (
	"context"
	"testing"

	"github.com/shuza/Autonoma/internal/sales/repository"
)

func TestCreateLeadRequiresCompanyName(t *testing.T) {
	t.Parallel()

	service := NewLeadService(&repository.LeadRepository{}, &repository.WorkflowRepository{})
	_, err := service.CreateLead(context.Background(), CreateLeadInput{})
	if err == nil {
		t.Fatalf("expected error for missing company name")
	}
}

func TestCreateLeadReturnsRepositoryErrors(t *testing.T) {
	t.Parallel()

	service := NewLeadService(&repository.LeadRepository{}, &repository.WorkflowRepository{})
	_, err := service.CreateLead(context.Background(), CreateLeadInput{})
	if err == nil {
		t.Fatalf("expected repository configuration error")
	}
}
