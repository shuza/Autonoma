package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/sales/service"
)

type leadCreator interface {
	CreateLead(ctx context.Context, input service.CreateLeadInput) (service.CreateLeadResult, error)
}

type createLeadRequest struct {
	CompanyName      string `json:"company_name"`
	Website          string `json:"website"`
	Source           string `json:"source"`
	ContactFirstName string `json:"contact_first_name"`
	ContactLastName  string `json:"contact_last_name"`
	ContactEmail     string `json:"contact_email"`
}

type createLeadResponse struct {
	Company  companyResponse  `json:"company"`
	Contact  contactResponse  `json:"contact"`
	Lead     leadResponse     `json:"lead"`
	Workflow workflowResponse `json:"workflow"`
}

type companyResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Website string `json:"website"`
}

type contactResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type leadResponse struct {
	ID        string            `json:"id"`
	CompanyID string            `json:"company_id"`
	Source    string            `json:"source"`
	Status    domain.LeadStatus `json:"status"`
}

type workflowResponse struct {
	ID     string                `json:"id"`
	LeadID string                `json:"lead_id"`
	Status domain.WorkflowStatus `json:"status"`
}

func handleCreateLead(creator leadCreator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if creator == nil {
			c.String(http.StatusServiceUnavailable, "lead service is not configured")
			return
		}

		var request createLeadRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.String(http.StatusBadRequest, "invalid request body")
			return
		}

		result, err := creator.CreateLead(c.Request.Context(), service.CreateLeadInput{
			CompanyName:      request.CompanyName,
			Website:          request.Website,
			Source:           request.Source,
			ContactFirstName: request.ContactFirstName,
			ContactLastName:  request.ContactLastName,
			ContactEmail:     request.ContactEmail,
		})
		if err != nil {
			writeCreateLeadError(c, err)
			return
		}

		c.JSON(http.StatusCreated, createLeadResponse{
			Company: companyResponse{
				ID:      result.Company.ID,
				Name:    result.Company.Name,
				Website: result.Company.Website,
			},
			Contact: contactResponse{
				ID:        result.Contact.ID,
				FirstName: result.Contact.FirstName,
				LastName:  result.Contact.LastName,
				Email:     result.Contact.Email,
			},
			Lead: leadResponse{
				ID:        result.Lead.ID,
				CompanyID: result.Company.ID,
				Source:    result.Lead.Source,
				Status:    result.Lead.Status,
			},
			Workflow: workflowResponse{
				ID:     result.Workflow.ID,
				LeadID: result.Workflow.LeadID,
				Status: result.Workflow.Status,
			},
		})
	}
}

func writeCreateLeadError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	if err.Error() == "company name is required" || err.Error() == "contact email is required" {
		status = http.StatusBadRequest
	}
	c.String(status, fmt.Sprintf("create lead: %v", err))
}
