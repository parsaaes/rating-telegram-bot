package database

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func New(fileName string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(fileName+"?_foreign_keys=on"), &gorm.Config{})
}