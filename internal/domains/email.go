package domains

import (
	"context"
	"time"
)

type Status string

const (
	Pending  Status = "pending"
	Retrying Status = "retrying"
	Sent     Status = "sent"
	Failed   Status = "failed"
)

func (s Status) IsValid() bool {
    switch s {
    case Pending, Retrying, Sent, Failed:
        return true
    default:
        return false
    }
}

type EmailEntity struct {
	ID        string
	IpAddress string
	Name      string
	ToEmail   string
	Type      EmailType
	Status    Status
	Project   *Project
	Attempt   int
	LastError *string
	SentAt    *time.Time
}

type EmailRepository interface {
	CreateEmail(ctx context.Context, email *EmailEntity) error
	GetEmailByID(ctx context.Context, id string) (*EmailEntity, error)
	GetEmailByIP(ctx context.Context, ip string) (*EmailEntity, error)
	GetEmails(ctx context.Context, page, limit int) ([]EmailEntity, int64, error)
	GetEmailsByStatus(ctx context.Context, status Status, page, limit int) ([]EmailEntity, int64, error)
	GetTodayEmails(ctx context.Context, page, limit int) ([]EmailEntity, int64, error)
	GetEmailsByLiveDemoRequest(ctx context.Context, typ string, page, limit int) ([]EmailEntity, int64, error)
	UpdateEmail(ctx context.Context, email *EmailEntity) error
}