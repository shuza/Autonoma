package domain

import "time"

type Company struct {
	ID        string
	Name      string
	Website   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
