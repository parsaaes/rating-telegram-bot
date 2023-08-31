package model

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	GroupID string `gorm:"uniqueIndex:unique_group_id_name"`
	Name string `gorm:"uniqueIndex:unique_group_id_name"`
	Icon string
}

type CategoryRepo interface {
	Create(category *Category) error
	Update(category *Category) error
	FindByName(name, groupID string) (*Category, error)
}

type SQLCategoryRepo struct {
	DB *gorm.DB
}

func (scr *SQLCategoryRepo) Create(category *Category) error {
	return scr.DB.Create(category).Error
}

func (scr *SQLCategoryRepo) Update(category *Category) error {
	return scr.DB.Updates(category).Error
}

func (scr *SQLCategoryRepo) FindByName(name, groupID string) (*Category, error)  {
	var category Category

	return &category, scr.DB.Where("name = ? and group_id = ?", name, groupID).First(&category).Error
}