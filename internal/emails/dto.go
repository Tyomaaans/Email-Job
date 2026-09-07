package emails

import (
	"time"

	"email-job/internal/domains"
)

// Requests

type SendMessageEmailRequest struct {
	Name    string `json:"name"    validate:"required,alphaspaceunicode"`
	Email   string `json:"email"   validate:"required,email"`
	Message string `json:"message" validate:"required"`
}

type SendLiveDemoRequest struct {
	Name    string          `json:"name"    validate:"required,alphaspaceunicode"`
	Email   string          `json:"email"   validate:"required,email"`
	Project domains.Project `json:"project" validate:"required,oneof=auth-session rate-limiter url-shorten ip-geolocation"`
	Message string          `json:"message" validate:"required"`
}

type SendLiveDemoManualRequest struct {
	TestingLink string `json:"testing_link" validate:"required,url"`
}

// Response

type EmailResponse struct {
	ID        string            `json:"id"`
	IpAddress string            `json:"ip_address"`
	Name      string            `json:"name"`
	ToEmail   string            `json:"to_email"`
	Type      domains.EmailType `json:"type"`
	Status    domains.Status    `json:"status"`
	Project   *domains.Project  `json:"project,omitempty"`
	Attempt   int               `json:"attempt"`
	LastError *string           `json:"last_error"`
	SentAt    *time.Time        `json:"sent_at"`
}

type PaginationMeta struct {
    CurrentPage int   `json:"current_page"`
    TotalPages  int   `json:"total_pages"`
    PageSize    int   `json:"page_size"`
    TotalData   int64 `json:"total_data"`
}

type PaginatedEmailResponse struct {
    Data []EmailResponse `json:"emails"`
    Meta PaginationMeta  `json:"meta"`
}