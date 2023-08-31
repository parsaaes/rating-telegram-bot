package model

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Rate struct {
	gorm.Model
	Rater  string `gorm:"uniqueIndex:unique_rater_item_id"`
	ItemID uint   `gorm:"uniqueIndex:unique_rater_item_id"`
	Rate   float64
	Item   Item
}

type RateToItem struct {
	Item     Item     `gorm:"embedded"`
	Category Category `gorm:"embedded"`
	Rate     Rate     `gorm:"embedded"`
}

type RateRepo interface {
	Save(rate *Rate) error
	List(groupID string) (map[Category]map[Item][]Rate, error)
}

type SQLRateRepo struct {
	DB *gorm.DB
}

func (srr *SQLRateRepo) Save(rate *Rate) error {
	return srr.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "rater"}, {Name: "item_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"rate", "updated_at"}),
	}).Create(rate).Error
}

func (srr *SQLRateRepo) List(groupID string) (map[Category]map[Item][]Rate, error) {
	var categories []Category

	err := srr.DB.Where("group_id = ?", groupID).Find(&categories).Error
	if err != nil {
		return nil, err
	}

	result := make(map[Category]map[Item][]Rate)

	for _, category := range categories {
		result[category] = make(map[Item][]Rate)

		var items []Item

		err := srr.DB.Where("category_id = ?", category.ID).Find(&items).Error
		if err != nil {
			return nil, err
		}

		for _, item := range items {
			result[category][item] = make([]Rate, 0)

			var rates []Rate

			err := srr.DB.Where("item_id = ?", item.ID).Order("rate desc").Find(&rates).Error
			if err != nil {
				return nil, err
			}

			result[category][item] = rates
		}
	}

	return result, nil
}
