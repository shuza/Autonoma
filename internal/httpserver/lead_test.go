package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shuza/Autonoma/internal/domain"
	"github.com/shuza/Autonoma/internal/platform/config"
	"github.com/shuza/Autonoma/internal/sales/service"
)

func newTestServer(creator leadCreator) *http.Server {
	return NewWithDependencies(config.Config{Host: "127.0.0.1", Port: "8080"}, creator)
}

type stubLeadCreator struct {
	result service.CreateLeadResult
	err    error
	input  service.CreateLeadInput
}

func (s *stubLeadCreator) CreateLead(_ context.Context, input service.CreateLeadInput) (service.CreateLeadResult, error) {
	s.input = input
	return s.result, s.err
}

func TestCreateLead(t *testing.T) {
	testCases := []struct {
		name           string
		body           string
		creator        leadCreator
		expectedStatus int
		assertResponse func(*testing.T, *httptest.ResponseRecorder, *stubLeadCreator)
	}{
		{
			name: "returns created",
			body: `{"company_name": "Acme","website":"https://acme.test","source":"inbound"}`,
			creator: &stubLeadCreator{
				result: service.CreateLeadResult{
					Company: domain.Company{
						ID:      "company-1",
						Name:    "Acme",
						Website: "http://acme.test",
					},
					Lead: domain.Lead{
						ID:        "lead-1",
						CompanyID: "company-1",
						Source:    "inbound",
						Status:    domain.LeadStatusNew,
					},
					Workflow: domain.Workflow{
						ID:     "workflow-1",
						LeadID: "lead-1",
						Status: domain.WorkflowStatusNew,
					},
				},
			},
			expectedStatus: http.StatusCreated,
			assertResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, creator *stubLeadCreator) {
				t.Helper()

				if creator.input.CompanyName != "Acme" {
					t.Fatalf("expected company name to be forwarded, got %q", creator.input.CompanyName)
				}

				var response createLeadResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatalf("expected valid json response, got error: %v", err)
				}

				if response.Lead.ID != "lead-1" || response.Workflow.ID != "workflow-1" {
					t.Fatalf("unexpected response payload: %+v", response)
				}

				if response.Company.ID != "company-1" || response.Lead.CompanyID != "company-1" {
					t.Fatalf("unexpected company linkage in response payload: %+v", response)
				}
			},
		},
		{
			name:           "reject invalid json",
			body:           "{",
			creator:        &stubLeadCreator{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "returns bad request for validation errors",
			body:           `{"company_name":""}`,
			creator:        &stubLeadCreator{err: errors.New("company name is required")},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "returns service unavailable when dependency missing",
			body:           `{"company_name":"Acme"}`,
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/v1/leads", bytes.NewBufferString(tc.body))
			recorder := httptest.NewRecorder()

			newTestServer(tc.creator).Handler.ServeHTTP(recorder, request)

			if recorder.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, recorder.Code)
			}

			if tc.assertResponse != nil {
				creator, ok := tc.creator.(*stubLeadCreator)
				if !ok {
					t.Fatalf("expected creator to be stubLeadCreator, got %T", tc.creator)
				}

				tc.assertResponse(t, recorder, creator)
			}
		})
	}
}
