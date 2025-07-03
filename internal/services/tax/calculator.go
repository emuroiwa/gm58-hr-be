package tax

import (
	"math"
	"gm58-hr-backend/internal/models"
	"gm58-hr-backend/internal/services/currency"
)

type TaxCalculator struct {
	currencyService *currency.CurrencyService
}

type TaxBracket struct {
	Min       float64
	Max       float64
	Rate      float64
	Deduction float64
}

func NewTaxCalculator(currencyService *currency.CurrencyService) *TaxCalculator {
	return &TaxCalculator{
		currencyService: currencyService,
	}
}

func (tc *TaxCalculator) GetMonthlyTaxBrackets() []TaxBracket {
	return []TaxBracket{
		{Min: 0, Max: 100.00, Rate: 0.00, Deduction: 0},
		{Min: 100.01, Max: 300.00, Rate: 0.20, Deduction: 20.00},
		{Min: 300.01, Max: 1000.00, Rate: 0.25, Deduction: 35.00},
		{Min: 1000.01, Max: 2000.00, Rate: 0.30, Deduction: 85.00},
		{Min: 2000.01, Max: 3000.00, Rate: 0.35, Deduction: 185.00},
		{Min: 3000.01, Max: math.Inf(1), Rate: 0.40, Deduction: 335.00},
	}
}

func (tc *TaxCalculator) CalculateMonthlyPAYE(grossSalary float64, employeeCurrency string) (float64, error) {
	if grossSalary <= 0 {
		return 0, nil
	}

	// Convert to USD for tax calculation if needed
	grossSalaryUSD := grossSalary
	if employeeCurrency != "USD" {
		convertedAmount, err := tc.currencyService.ConvertAmount(grossSalary, employeeCurrency, "USD")
		if err != nil {
			return 0, err
		}
		grossSalaryUSD = convertedAmount
	}

	brackets := tc.GetMonthlyTaxBrackets()
	var tax float64

	for _, bracket := range brackets {
		if grossSalaryUSD >= bracket.Min && (grossSalaryUSD <= bracket.Max || math.IsInf(bracket.Max, 1)) {
			tax = grossSalaryUSD*bracket.Rate - bracket.Deduction
			if tax < 0 {
				tax = 0
			}
			break
		}
	}

	// Convert tax back to employee currency if needed
	if employeeCurrency != "USD" {
		convertedTax, err := tc.currencyService.ConvertAmount(tax, "USD", employeeCurrency)
		if err != nil {
			return 0, err
		}
		return convertedTax, nil
	}

	return tax, nil
}

func (tc *TaxCalculator) CalculateAidsLevy(payeeTax float64) float64 {
	return payeeTax * 0.03
}

func (tc *TaxCalculator) CalculateNSSAContribution(grossSalary float64, employeeCurrency string) (float64, error) {
	// NSSA contribution is 3% of gross salary
	nssaRate := 0.03
	contribution := grossSalary * nssaRate

	return contribution, nil
}

func (tc *TaxCalculator) CalculatePensionContribution(grossSalary, pensionRate float64) float64 {
	if pensionRate <= 0 {
		return 0
	}
	return grossSalary * (pensionRate / 100)
}

func (tc *TaxCalculator) CalculateYTDTax(employee models.Employee, currentYear int) (float64, error) {
	// Calculate year-to-date tax for the employee
	return 0, nil // Implement based on your requirements
}
