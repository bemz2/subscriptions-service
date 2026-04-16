package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const monthYearLayout = "01-2006"

// ParseMonthYear parses a "MM-YYYY" string into a time.Time (first day of the month, UTC).
func ParseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse(monthYearLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q, expected MM-YYYY format", s)
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

// FormatMonthYear formats a time.Time as "MM-YYYY".
func FormatMonthYear(t time.Time) string {
	return t.Format(monthYearLayout)
}

func FormatMonthYearPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := FormatMonthYear(*t)
	return &s
}

// CreateSubscriptionRequest is the body for POST /subscriptions.
type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"` // MM-YYYY
	EndDate     *string   `json:"end_date"`   // MM-YYYY, optional
}

// UpdateSubscriptionRequest is the body for PUT /subscriptions/:id.
type UpdateSubscriptionRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	StartDate   string  `json:"start_date"` // MM-YYYY
	EndDate     *string `json:"end_date"`   // MM-YYYY, optional
}

// SubscriptionResponse is returned for all single/list subscription endpoints.
type SubscriptionResponse struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"` // MM-YYYY
	EndDate     *string   `json:"end_date"`   // MM-YYYY, optional
}

// SumResponse is returned for GET /subscriptions/sum.
type SumResponse struct {
	Total int `json:"total"`
}
