package models

import (
	"time"

	"gorm.io/gorm"
)

type PayrollPeriod struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	CompanyID   uint       `json:"company_id"`
	Company     Company    `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	Year        int        `json:"year"`
	Month       int        `json:"month"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	Status      string     `json:"status" gorm:"default:'draft'"` // draft, processing, processed, approved, paid
	Description string     `json:"description"`
	ProcessedAt *time.Time `json:"processed_at"`
	ProcessedBy *uint      `json:"processed_by"`
	ApprovedAt  *time.Time `json:"approved_at"`
	ApprovedBy  *uint      `json:"approved_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relationships
	Payslips []Payslip `json:"payslips,omitempty" gorm:"foreignKey:PayrollPeriodID"`
}

type Payslip struct {
	ID              uint          `json:"id" gorm:"primaryKey"`
	CompanyID       uint          `json:"company_id"`
	Company         Company       `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	EmployeeID      uint          `json:"employee_id"`
	PayrollPeriodID uint          `json:"payroll_period_id"`
	Employee        Employee      `json:"employee" gorm:"foreignKey:EmployeeID"`
	PayrollPeriod   PayrollPeriod `json:"payroll_period" gorm:"foreignKey:PayrollPeriodID"`

	// Currency Information
	CurrencyID   uint     `json:"currency_id"`
	Currency     Currency `json:"currency" gorm:"foreignKey:CurrencyID"`
	ExchangeRate float64  `json:"exchange_rate" gorm:"type:decimal(15,6)"` // Rate to base currency

	// Earnings (in employee's currency)
	BasicSalary   float64 `json:"basic_salary" gorm:"type:decimal(15,2)"`
	Overtime      float64 `json:"overtime" gorm:"type:decimal(15,2)"`
	Allowances    float64 `json:"allowances" gorm:"type:decimal(15,2)"`
	Bonus         float64 `json:"bonus" gorm:"type:decimal(15,2)"`
	Commission    float64 `json:"commission" gorm:"type:decimal(15,2)"`
	OtherEarnings float64 `json:"other_earnings" gorm:"type:decimal(15,2)"`
	TotalEarnings float64 `json:"total_earnings" gorm:"type:decimal(15,2)"`

	// Deductions (in employee's currency)
	PayeeTax            float64 `json:"payee_tax" gorm:"type:decimal(15,2)"`
	AidsLevy            float64 `json:"aids_levy" gorm:"type:decimal(15,2)"`
	NSSAContribution    float64 `json:"nssa_contribution" gorm:"type:decimal(15,2)"`
	PensionContribution float64 `json:"pension_contribution" gorm:"type:decimal(15,2)"`
	MedicalAid          float64 `json:"medical_aid" gorm:"type:decimal(15,2)"`
	UnionDues           float64 `json:"union_dues" gorm:"type:decimal(15,2)"`
	LoanDeductions      float64 `json:"loan_deductions" gorm:"type:decimal(15,2)"`
	OtherDeductions     float64 `json:"other_deductions" gorm:"type:decimal(15,2)"`
	TotalDeductions     float64 `json:"total_deductions" gorm:"type:decimal(15,2)"`

	// Net Pay (in employee's currency)
	NetPay float64 `json:"net_pay" gorm:"type:decimal(15,2)"`

	// Base Currency Amounts (for reporting)
	TotalEarningsBase   float64 `json:"total_earnings_base" gorm:"type:decimal(15,2)"`
	TotalDeductionsBase float64 `json:"total_deductions_base" gorm:"type:decimal(15,2)"`
	NetPayBase          float64 `json:"net_pay_base" gorm:"type:decimal(15,2)"`

	// Working Days
	WorkingDays int `json:"working_days"`
	DaysWorked  int `json:"days_worked"`
	DaysAbsent  int `json:"days_absent"`

	// Status
	Status           string     `json:"status" gorm:"default:'generated'"` // generated, approved, paid
	PaymentReference string     `json:"payment_reference"`
	PaymentDate      *time.Time `json:"payment_date"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Allowance struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CompanyID   uint           `json:"company_id"`
	Company     Company        `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	EmployeeID  uint           `json:"employee_id"`
	Employee    Employee       `json:"employee" gorm:"foreignKey:EmployeeID"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Amount      float64        `json:"amount" gorm:"type:decimal(15,2)"`
	CurrencyID  uint           `json:"currency_id"`
	Currency    Currency       `json:"currency" gorm:"foreignKey:CurrencyID"`
	IsFixed     bool           `json:"is_fixed" gorm:"default:true"`
	Percentage  float64        `json:"percentage" gorm:"type:decimal(5,2)"`
	IsTaxable   bool           `json:"is_taxable" gorm:"default:true"`
	IsRecurring bool           `json:"is_recurring" gorm:"default:true"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     *time.Time     `json:"end_date"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type Deduction struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CompanyID   uint           `json:"company_id"`
	Company     Company        `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	EmployeeID  uint           `json:"employee_id"`
	Employee    Employee       `json:"employee" gorm:"foreignKey:EmployeeID"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Amount      float64        `json:"amount" gorm:"type:decimal(15,2)"`
	CurrencyID  uint           `json:"currency_id"`
	Currency    Currency       `json:"currency" gorm:"foreignKey:CurrencyID"`
	IsFixed     bool           `json:"is_fixed" gorm:"default:true"`
	Percentage  float64        `json:"percentage" gorm:"type:decimal(5,2)"`
	IsStatutory bool           `json:"is_statutory" gorm:"default:false"`
	IsRecurring bool           `json:"is_recurring" gorm:"default:true"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     *time.Time     `json:"end_date"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}
