package models

import (
	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	SenderAddress    string  `gorm:"type:varchar(42);not null"`
	RecipientAddress string  `gorm:"type:varchar(42);not null"`
	AmountETH        string  `gorm:"type:decimal(36,18);not null"`
	AmountUSD        *string `gorm:"type:decimal(20,6)"`
	Status           string  `gorm:"type:varchar(20);default:'completed'"`
	TransactionHash  string  `gorm:"type:varchar(66);uniqueIndex"`
}

func (t Transaction) TableName() string {
	return "transaction"
}
