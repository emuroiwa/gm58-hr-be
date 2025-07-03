package handlers

import (
	"net/http"
	"strconv"
	"gm58-hr-backend/internal/models"
	"gm58-hr-backend/internal/services/currency"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CurrencyHandler struct {
	db              *gorm.DB
	currencyService *currency.CurrencyService
}

func NewCurrencyHandler(db *gorm.DB, currencyService *currency.CurrencyService) *CurrencyHandler {
	return &CurrencyHandler{
		db:              db,
		currencyService: currencyService,
	}
}

func (ch *CurrencyHandler) GetCurrencies(c *gin.Context) {
	currencies, err := ch.currencyService.GetSupportedCurrencies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch currencies"})
		return
	}

	c.JSON(http.StatusOK, currencies)
}

func (ch *CurrencyHandler) CreateCurrency(c *gin.Context) {
	var currency models.Currency
	if err := c.ShouldBindJSON(&currency); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ch.db.Create(&currency).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create currency"})
		return
	}

	c.JSON(http.StatusCreated, currency)
}

func (ch *CurrencyHandler) GetExchangeRate(c *gin.Context) {
	fromCurrency := c.Query("from")
	toCurrency := c.Query("to")

	if fromCurrency == "" || toCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both 'from' and 'to' currencies are required"})
		return
	}

	rate, err := ch.currencyService.GetExchangeRate(fromCurrency, toCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get exchange rate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"from_currency": fromCurrency,
		"to_currency":   toCurrency,
		"rate":          rate,
	})
}

func (ch *CurrencyHandler) ConvertAmount(c *gin.Context) {
	amountStr := c.Query("amount")
	fromCurrency := c.Query("from")
	toCurrency := c.Query("to")

	if amountStr == "" || fromCurrency == "" || toCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount, from, and to currencies are required"})
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}

	convertedAmount, err := ch.currencyService.ConvertAmount(amount, fromCurrency, toCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert amount"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"original_amount":  amount,
		"from_currency":    fromCurrency,
		"to_currency":      toCurrency,
		"converted_amount": convertedAmount,
	})
}

func (ch *CurrencyHandler) UpdateExchangeRates(c *gin.Context) {
	if err := ch.currencyService.UpdateExchangeRates(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exchange rates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exchange rates updated successfully"})
}

func (ch *CurrencyHandler) GetExchangeRateHistory(c *gin.Context) {
	fromCurrency := c.Query("from")
	toCurrency := c.Query("to")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))

	if fromCurrency == "" || toCurrency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both 'from' and 'to' currencies are required"})
		return
	}

	var rates []models.ExchangeRate
	err := ch.db.Joins("FromCurrency").Joins("ToCurrency").
		Where("FromCurrency.code = ? AND ToCurrency.code = ?", fromCurrency, toCurrency).
		Order("effective_date DESC").
		Limit(limit).
		Find(&rates).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exchange rate history"})
		return
	}

	c.JSON(http.StatusOK, rates)
}
