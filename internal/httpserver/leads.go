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
	CompanyName string `json:"company_name"`
	Website     string `json:"website"`
	Source      string `json:"source"`
}

type createLeadResponse struct {
	Lead     leadResponse     `json:"lead"`
	Workflow workflowResponse `json:"workflow"`
}

type leadResponse struct {
	ID          string            `json:"id"`
	CompanyName string            `json:"company_name"`
	Website     string            `json:"website"`
	Source      string            `json:"source"`
	Status      domain.LeadStatus `json:"status"`
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
			CompanyName: request.CompanyName,
			Website:     request.Website,
			Source:      request.Source,
		})
		if err != nil {
			writeCreateLeadError(c, err)
			return
		}

		c.JSON(http.StatusCreated, createLeadResponse{
			Lead: leadResponse{
				ID:          result.Lead.ID,
				CompanyName: result.Lead.CompanyName,
				Website:     result.Lead.Website,
				Source:      result.Lead.Source,
				Status:      result.Lead.Status,
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
	if err.Error() == "company name is required" {
		status = http.StatusBadRequest
	}
	c.String(status, fmt.Sprintf("create lead: %v", err))
}
