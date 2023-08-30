package model

import "gorm.io/gorm"

type Item struct {
	gorm.Model
	Title string `gorm:"uniqueIndex:unique_category_id_title"`
	CategoryID int `gorm:"uniqueIndex:unique_category_id_title"`
	Category Category
}
