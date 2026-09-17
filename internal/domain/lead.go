package domain

import "time"

type LeadStatus string

const (
	LeadStatusNew LeadStatus = "NEW"
)

type Lead struct {
	ID          string
	CompanyName string
	Website     string
	Source      string
	Status      LeadStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
