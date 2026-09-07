package emails

import (
	"context"
	"time"

	"gorm.io/gorm"

	"email-job/internal/domains"
	"email-job/internal/infrastructures/sqlite"
	"email-job/pkg"
)

type emailRepository struct {
	db *gorm.DB
}

func NewEmailRepository(db *gorm.DB) domains.EmailRepository {
	return &emailRepository{
		db: db,
	}
}

// Email Job

func (r *emailRepository) CreateEmail(ctx context.Context, email *domains.EmailEntity) error {
	storage := ToEmailStorage(email)

	if err := r.db.WithContext(ctx).
		Create(storage).
		Error; err != nil {
			return pkg.HandleDBError(err)
		}

	return nil
}

func (r *emailRepository) GetEmailByID(ctx context.Context, id string) (*domains.EmailEntity, error) {
	var storage sqlite.Email

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&storage).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToEmailEntity(&storage), nil
}

func (r *emailRepository) GetEmailByIP(ctx context.Context, ip string) (*domains.EmailEntity, error) {
	var storage sqlite.Email

	err := r.db.WithContext(ctx).
		Where("ip_address = ?", ip).
		First(&storage).Error

	if err != nil {
		return nil, pkg.HandleDBError(err)
	}

	return ToEmailEntity(&storage), nil
}

func (r *emailRepository) GetEmails(ctx context.Context, page, limit int) ([]domains.EmailEntity, int64, error) {
	var storages []sqlite.Email
	var totalData int64

	if err := r.db.WithContext(ctx).Model(&sqlite.Email{}).Count(&totalData).Error; err != nil {
		return nil, 0, pkg.HandleDBError(err)
	}

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	if err := r.db.WithContext(ctx).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&storages).Error; err != nil {
		return nil, 0, pkg.HandleDBError(err)
	}

	return ToEmailListEntities(storages), totalData, nil
}

func (r *emailRepository) GetEmailsByStatus(ctx context.Context, status domains.Status, page, limit int) ([]domains.EmailEntity, int64, error) {
	var storages []sqlite.Email
	var totalData int64

	query := r.db.WithContext(ctx).
		Model(&sqlite.Email{}).
		Where("status = ?", status)

	if err := query.Count(&totalData).Error; err != nil {
		return nil, 0, pkg.HandleDBError(err)
	}

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	if err := query.
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&storages).Error; err != nil {
		return nil, 0, pkg.HandleDBError(err)
	}

	return ToEmailListEntities(storages), totalData, nil
}

func (r *emailRepository) GetTodayEmails(ctx context.Context, page, limit int) ([]domains.EmailEntity, int64, error) {
	var storages []sqlite.Email
	var totalData int64

	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-1 * time.Nanosecond)

	query := r.db.WithContext(ctx).
		Model(&sqlite.Email{}).
		Where("created_at BETWEEN ? AND ?", startOfDay, endOfDay)

	if err := query.Count(&totalData).Error; err != nil {
		return nil, 0, pkg.HandleDBError(err)
	}

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	if err := query.
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&storages).Error; err != nil {
		return nil, 0, pkg.HandleDBError(err)
	}

	return ToEmailListEntities(storages), totalData, nil
}

func (r *emailRepository) GetEmailsByLiveDemoRequest(ctx context.Context, typ string, page, limit int) ([]domains.EmailEntity, int64, error) {
	var storages []sqlite.Email
	var totalData int64

	query := r.db.WithContext(ctx).
		Model(&sqlite.Email{}).
		Where("type = ?", typ)

	if err := query.Count(&totalData).Error; err != nil {
		return nil, 0, pkg.HandleDBError(err)
	}

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	if err := query.
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&storages).Error; err != nil {
		return nil, 0, pkg.HandleDBError(err)
	}

	return ToEmailListEntities(storages), totalData, nil
}

func (r *emailRepository) UpdateEmail(ctx context.Context, email *domains.EmailEntity) error {
	storage := ToEmailStorage(email)

	if err := r.db.WithContext(ctx).
		Model(&sqlite.Email{}).
		Where("id = ?", storage.ID).
		Updates(storage).
		Error; err != nil {
			return pkg.HandleDBError(err)
		}

	return nil
}