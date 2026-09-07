package sqlite

import (
	"time"

	"email-job/internal/domains"
)

type Email struct {
	ID        string            `gorm:"primaryKey"`
	IpAddress string            `gorm:"type:varchar(255);not null"`
	Name      string            `gorm:"type:varchar(255);not null"`
	ToEmail   string            `gorm:"type:varchar(255);not null"`
	Type      domains.EmailType `gorm:"type:varchar(255);not null"`
	Status    domains.Status    `gorm:"type:varchar(100);not null"`
	Project   *domains.Project  `gorm:"type:varchar(100)"`
	Attempt   int               `gorm:"type:smallint;not null"`
	LastError *string           `gorm:"type:text"`
	SentAt    *time.Time        `gorm:"index"`
	CreatedAt time.Time         `gorm:"not null"`
	UpdatedAt time.Time         `gorm:"not null;index"`
}