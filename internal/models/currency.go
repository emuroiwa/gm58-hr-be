package models

import (
	"time"
	"gorm.io/gorm"
)

type Currency struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Code        string `json:"code" gorm:"unique;not null;size:3"` // USD, ZWL, ZAR, etc.
	Name        string `json:"name" gorm:"not null"`               // US Dollar, Local Dollar, etc.
	Symbol      string `json:"symbol" gorm:"size:5"`               // $, Z$, R, etc.
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	IsBaseCurrency bool `json:"is_base_currency" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type ExchangeRate struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	FromCurrencyID uint      `json:"from_currency_id"`
	ToCurrencyID   uint      `json:"to_currency_id"`
	FromCurrency   Currency  `json:"from_currency" gorm:"foreignKey:FromCurrencyID"`
	ToCurrency     Currency  `json:"to_currency" gorm:"foreignKey:ToCurrencyID"`
	Rate           float64   `json:"rate" gorm:"type:decimal(15,6)"`
	EffectiveDate  time.Time `json:"effective_date"`
	Source         string    `json:"source"` // API, manual, etc.
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
