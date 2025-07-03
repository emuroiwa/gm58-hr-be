package payroll

import (
	"fmt"
	"time"
	"gm58-hr-backend/internal/models"
	"gm58-hr-backend/internal/services/currency"
	"gm58-hr-backend/internal/services/tax"

	"gorm.io/gorm"
)

type PayrollProcessor struct {
	db              *gorm.DB
	taxCalculator   *tax.TaxCalculator
	currencyService *currency.CurrencyService
}

func NewPayrollProcessor(db *gorm.DB, currencyService *currency.CurrencyService) *PayrollProcessor {
	return &PayrollProcessor{
		db:              db,
		taxCalculator:   tax.NewTaxCalculator(currencyService),
		currencyService: currencyService,
	}
}

func (pp *PayrollProcessor) ProcessPayroll(periodID uint) error {
	var period models.PayrollPeriod
	if err := pp.db.First(&period, periodID).Error; err != nil {
		return fmt.Errorf("payroll period not found: %w", err)
	}

	if period.Status != "draft" {
		return fmt.Errorf("payroll period is not in draft status")
	}

	// Update status to processing
	period.Status = "processing"
	pp.db.Save(&period)

	var employees []models.Employee
	if err := pp.db.Preload("Currency").Preload("Position").Preload("Department").
		Where("is_active = ? AND employment_status = ?", true, "active").
		Find(&employees).Error; err != nil {
		return fmt.Errorf("failed to fetch employees: %w", err)
	}

	for _, employee := range employees {
		if err := pp.processEmployeePayroll(employee, period); err != nil {
			fmt.Printf("Failed to process payroll for employee %s: %v\n", employee.EmployeeNumber, err)
			continue
		}
	}

	// Update period status
	now := time.Now()
	period.Status = "processed"
	period.ProcessedAt = &now
	if err := pp.db.Save(&period).Error; err != nil {
		return fmt.Errorf("failed to update period status: %w", err)
	}

	return nil
}

func (pp *PayrollProcessor) processEmployeePayroll(employee models.Employee, period models.PayrollPeriod) error {
	// Check if payslip already exists
	var existingPayslip models.Payslip
	if err := pp.db.Where("employee_id = ? AND payroll_period_id = ?", 
		employee.ID, period.ID).First(&existingPayslip).Error; err == nil {
		return fmt.Errorf("payslip already exists for employee %s", employee.EmployeeNumber)
	}

	// Get exchange rate for employee currency to base currency
	baseCurrency, err := pp.currencyService.GetBaseCurrency()
	if err != nil {
		return fmt.Errorf("failed to get base currency: %w", err)
	}

	exchangeRate := 1.0
	if employee.Currency.Code != baseCurrency.Code {
		rate, err := pp.currencyService.GetExchangeRate(employee.Currency.Code, baseCurrency.Code)
		if err != nil {
			return fmt.Errorf("failed to get exchange rate: %w", err)
		}
		exchangeRate = rate
	}

	// Calculate earnings
	basicSalary := employee.BasicSalary
	
	// Get allowances for the employee
	allowances, err := pp.calculateAllowances(employee.ID, employee.Currency.Code)
	if err != nil {
		return fmt.Errorf("failed to calculate allowances: %w", err)
	}

	// Calculate overtime
	overtime := pp.calculateOvertime(employee.ID, period)

	// Calculate bonus
	bonus := pp.calculateBonus(employee.ID, period)

	totalEarnings := basicSalary + allowances + overtime + bonus

	// Calculate deductions
	payeeTax, err := pp.taxCalculator.CalculateMonthlyPAYE(totalEarnings, employee.Currency.Code)
	if err != nil {
		return fmt.Errorf("failed to calculate PAYE: %w", err)
	}

	aidsLevy := pp.taxCalculator.CalculateAidsLevy(payeeTax)
	
	nssaContribution, err := pp.taxCalculator.CalculateNSSAContribution(totalEarnings, employee.Currency.Code)
	if err != nil {
		return fmt.Errorf("failed to calculate NSSA: %w", err)
	}

	// Get other deductions
	otherDeductions, err := pp.calculateDeductions(employee.ID, employee.Currency.Code)
	if err != nil {
		return fmt.Errorf("failed to calculate deductions: %w", err)
	}

	totalDeductions := payeeTax + aidsLevy + nssaContribution + otherDeductions
	netPay := totalEarnings - totalDeductions

	// Convert to base currency for reporting
	totalEarningsBase := totalEarnings * exchangeRate
	totalDeductionsBase := totalDeductions * exchangeRate
	netPayBase := netPay * exchangeRate

	// Get working days
	workingDays := pp.calculateWorkingDays(period.StartDate, period.EndDate)
	daysWorked := pp.getDaysWorked(employee.ID, period)

	// Create payslip
	payslip := models.Payslip{
		EmployeeID:          employee.ID,
		PayrollPeriodID:     period.ID,
		CurrencyID:          employee.CurrencyID,
		ExchangeRate:        exchangeRate,
		BasicSalary:         basicSalary,
		Overtime:            overtime,
		Allowances:          allowances,
		Bonus:               bonus,
		TotalEarnings:       totalEarnings,
		PayeeTax:            payeeTax,
		AidsLevy:            aidsLevy,
		NSSAContribution:    nssaContribution,
		OtherDeductions:     otherDeductions,
		TotalDeductions:     totalDeductions,
		NetPay:              netPay,
		TotalEarningsBase:   totalEarningsBase,
		TotalDeductionsBase: totalDeductionsBase,
		NetPayBase:          netPayBase,
		WorkingDays:         workingDays,
		DaysWorked:          daysWorked,
		DaysAbsent:          workingDays - daysWorked,
		Status:              "generated",
	}

	return pp.db.Create(&payslip).Error
}

func (pp *PayrollProcessor) calculateAllowances(employeeID uint, currency string) (float64, error) {
	var allowances []models.Allowance
	err := pp.db.Preload("Currency").Where("employee_id = ? AND is_active = ? AND is_recurring = ?", 
		employeeID, true, true).Find(&allowances).Error
	if err != nil {
		return 0, err
	}

	total := 0.0
	for _, allowance := range allowances {
		amount := allowance.Amount
		
		// Convert to employee currency if different
		if allowance.Currency.Code != currency {
			convertedAmount, err := pp.currencyService.ConvertAmount(amount, allowance.Currency.Code, currency)
			if err != nil {
				continue
			}
			amount = convertedAmount
		}
		
		total += amount
	}

	return total, nil
}

func (pp *PayrollProcessor) calculateDeductions(employeeID uint, currency string) (float64, error) {
	var deductions []models.Deduction
	err := pp.db.Preload("Currency").Where("employee_id = ? AND is_active = ? AND is_recurring = ?", 
		employeeID, true, true).Find(&deductions).Error
	if err != nil {
		return 0, err
	}

	total := 0.0
	for _, deduction := range deductions {
		amount := deduction.Amount
		
		// Convert to employee currency if different
		if deduction.Currency.Code != currency {
			convertedAmount, err := pp.currencyService.ConvertAmount(amount, deduction.Currency.Code, currency)
			if err != nil {
				continue
			}
			amount = convertedAmount
		}
		
		total += amount
	}

	return total, nil
}

func (pp *PayrollProcessor) calculateOvertime(employeeID uint, period models.PayrollPeriod) float64 {
	return 0
}

func (pp *PayrollProcessor) calculateBonus(employeeID uint, period models.PayrollPeriod) float64 {
	return 0
}

func (pp *PayrollProcessor) calculateWorkingDays(startDate, endDate time.Time) int {
	days := 0
	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			days++
		}
	}
	return days
}

func (pp *PayrollProcessor) getDaysWorked(employeeID uint, period models.PayrollPeriod) int {
	return pp.calculateWorkingDays(period.StartDate, period.EndDate)
}

func (pp *PayrollProcessor) ApprovePayroll(periodID uint, approverID uint) error {
	var period models.PayrollPeriod
	if err := pp.db.First(&period, periodID).Error; err != nil {
		return fmt.Errorf("payroll period not found: %w", err)
	}

	if period.Status != "processed" {
		return fmt.Errorf("payroll period must be processed before approval")
	}

	now := time.Now()
	period.Status = "approved"
	period.ApprovedAt = &now
	period.ApprovedBy = &approverID

	return pp.db.Save(&period).Error
}

func (pp *PayrollProcessor) GetPayrollSummary(periodID uint) (map[string]interface{}, error) {
	var payslips []models.Payslip
	err := pp.db.Where("payroll_period_id = ?", periodID).Find(&payslips).Error
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"total_employees":     len(payslips),
		"total_earnings":      0.0,
		"total_deductions":    0.0,
		"total_net_pay":       0.0,
		"total_paye_tax":      0.0,
		"total_nssa":          0.0,
		"currency_breakdown":  make(map[string]interface{}),
	}

	currencyBreakdown := make(map[string]map[string]float64)

	for _, payslip := range payslips {
		summary["total_earnings"] = summary["total_earnings"].(float64) + payslip.TotalEarningsBase
		summary["total_deductions"] = summary["total_deductions"].(float64) + payslip.TotalDeductionsBase
		summary["total_net_pay"] = summary["total_net_pay"].(float64) + payslip.NetPayBase
		summary["total_paye_tax"] = summary["total_paye_tax"].(float64) + payslip.PayeeTax*payslip.ExchangeRate
		summary["total_nssa"] = summary["total_nssa"].(float64) + payslip.NSSAContribution*payslip.ExchangeRate

		// Track by currency
		var currency models.Currency
		pp.db.First(&currency, payslip.CurrencyID)
		
		if currencyBreakdown[currency.Code] == nil {
			currencyBreakdown[currency.Code] = make(map[string]float64)
		}
		
		currencyBreakdown[currency.Code]["total_earnings"] += payslip.TotalEarnings
		currencyBreakdown[currency.Code]["total_net_pay"] += payslip.NetPay
		currencyBreakdown[currency.Code]["employee_count"] += 1
	}

	summary["currency_breakdown"] = currencyBreakdown
	return summary, nil
}
