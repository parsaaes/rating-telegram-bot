package model

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	GroupID string `gorm:"uniqueIndex:unique_group_id_name"`
	Name string `gorm:"uniqueIndex:unique_group_id_name"`
}

type CategoryRepo interface {
	Create(category *Category) error
}

type SQLCategoryRepo struct {
	DB *gorm.DB
}

func (scr *SQLCategoryRepo) Create(category *Category) error {
	return scr.DB.Create(category).Error
}