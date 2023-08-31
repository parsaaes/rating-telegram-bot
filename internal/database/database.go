package database

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func New(fileName string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(fileName), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.Exec("PRAGMA foreign_keys = ON", nil).Error; err != nil {
		return nil, err
	}

	return db, nil
}