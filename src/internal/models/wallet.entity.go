package models

import (
	"gorm.io/gorm"
)

type Wallet struct {
	gorm.Model
	Address string  `gorm:"type:varchar(42);uniqueIndex;not null"`
	Balance string  `gorm:"type:decimal(36,18);not null"`
	Email   *string `gorm:"type:varchar(255)"`
}

func (W Wallet) TableName() string {
	return "wallets"
}
