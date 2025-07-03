package currency

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"gm58-hr-backend/internal/models"

	"gorm.io/gorm"
)

type CurrencyService struct {
	db     *gorm.DB
	apiKey string
	apiURL string
}

type ExchangeRateResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

func NewCurrencyService(db *gorm.DB, apiKey, apiURL string) *CurrencyService {
	return &CurrencyService{
		db:     db,
		apiKey: apiKey,
		apiURL: apiURL,
	}
}

func (cs *CurrencyService) GetExchangeRate(fromCurrency, toCurrency string) (float64, error) {
	// First try to get from database (cached rates)
	var exchangeRate models.ExchangeRate
	err := cs.db.Joins("FromCurrency").Joins("ToCurrency").
		Where("FromCurrency.code = ? AND ToCurrency.code = ? AND effective_date >= ?",
			fromCurrency, toCurrency, time.Now().AddDate(0, 0, -1)).
		Order("effective_date DESC").
		First(&exchangeRate).Error

	if err == nil {
		return exchangeRate.Rate, nil
	}

	// If not found or outdated, fetch from API
	rate, err := cs.fetchExchangeRateFromAPI(fromCurrency, toCurrency)
	if err != nil {
		return 0, err
	}

	// Save to database for caching
	go cs.saveExchangeRate(fromCurrency, toCurrency, rate)

	return rate, nil
}

func (cs *CurrencyService) fetchExchangeRateFromAPI(fromCurrency, toCurrency string) (float64, error) {
	url := fmt.Sprintf("%s%s", cs.apiURL, fromCurrency)
	
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var response ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, err
	}

	rate, exists := response.Rates[toCurrency]
	if !exists {
		return 0, fmt.Errorf("exchange rate not found for %s to %s", fromCurrency, toCurrency)
	}

	return rate, nil
}

func (cs *CurrencyService) saveExchangeRate(fromCurrency, toCurrency string, rate float64) {
	var fromCurr, toCurr models.Currency
	
	cs.db.Where("code = ?", fromCurrency).First(&fromCurr)
	cs.db.Where("code = ?", toCurrency).First(&toCurr)

	exchangeRate := models.ExchangeRate{
		FromCurrencyID: fromCurr.ID,
		ToCurrencyID:   toCurr.ID,
		Rate:          rate,
		EffectiveDate: time.Now(),
		Source:        "API",
	}

	cs.db.Create(&exchangeRate)
}

func (cs *CurrencyService) ConvertAmount(amount float64, fromCurrency, toCurrency string) (float64, error) {
	if fromCurrency == toCurrency {
		return amount, nil
	}

	rate, err := cs.GetExchangeRate(fromCurrency, toCurrency)
	if err != nil {
		return 0, err
	}

	return amount * rate, nil
}

func (cs *CurrencyService) GetSupportedCurrencies() ([]models.Currency, error) {
	var currencies []models.Currency
	err := cs.db.Where("is_active = ?", true).Find(&currencies).Error
	return currencies, err
}

func (cs *CurrencyService) GetBaseCurrency() (*models.Currency, error) {
	var currency models.Currency
	err := cs.db.Where("is_base_currency = ?", true).First(&currency).Error
	return &currency, err
}

func (cs *CurrencyService) UpdateExchangeRates() error {
	baseCurrency, err := cs.GetBaseCurrency()
	if err != nil {
		return err
	}

	currencies, err := cs.GetSupportedCurrencies()
	if err != nil {
		return err
	}

	for _, currency := range currencies {
		if currency.Code == baseCurrency.Code {
			continue
		}

		rate, err := cs.fetchExchangeRateFromAPI(baseCurrency.Code, currency.Code)
		if err != nil {
			continue // Skip failed rates
		}

		cs.saveExchangeRate(baseCurrency.Code, currency.Code, rate)
	}

	return nil
}
