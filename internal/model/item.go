package model

import (
	"gorm.io/gorm"
)

type Item struct {
	gorm.Model
	Title      string `gorm:"uniqueIndex:unique_category_id_title"`
	CategoryID uint   `gorm:"uniqueIndex:unique_category_id_title"`
	Category   Category
}

type ItemRepo interface {
	Create(item *Item) error
	FindIDByTitleAndGroupID(title, groupID string) (uint, error)
}

type SQLItemRepo struct {
	DB *gorm.DB
}

func (sir *SQLItemRepo) Create(item *Item) error {
	return sir.DB.Create(item).Error
}

func (sir *SQLItemRepo) FindIDByTitleAndGroupID(title, groupID string) (uint, error) {
	var res uint

	return res, sir.DB.
		Model(&Item{}).
		Select("items.id").
		Joins(joinCategories()).
		Where("items.title = ? and categories.group_id = ?", title, groupID).
		Scan(&res).
		Error
}

func joinCategories() string {
	return "join categories on items.category_id = categories.id"
}
