package domain

import "time"

type Contact struct {
	ID        string
	CompanyID string
	FirstName string
	LastName  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
