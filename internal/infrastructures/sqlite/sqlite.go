package sqlite

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/glebarez/sqlite"
)

func NewSQLiteDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to connect %w", err)
	}

	if err := db.AutoMigrate(
		&Email{},
	); err != nil {
		return nil, fmt.Errorf("sqlite: failed to migrate %w", err)
	}

	return db, nil
}