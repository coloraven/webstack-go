package model

import (
	"gorm.io/gorm"
)

// BeforeCreate StSite 添加前的回调
func (s *StSite) BeforeCreate(tx *gorm.DB) error {
	if !tx.Migrator().HasIndex(&StSite{}, "idx_category_url") {
		tx.Migrator().CreateIndex(&StSite{}, "idx_category_url")
	}
	return nil
}