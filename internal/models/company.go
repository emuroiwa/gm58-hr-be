package models

import (
	"time"

	"gorm.io/gorm"
)

type Company struct {
	ID             uint     `json:"id" gorm:"primaryKey"`
	Name           string   `json:"name" gorm:"unique;not null"`
	Code           string   `json:"code" gorm:"unique;not null;size:10"` // Unique company code/identifier
	Email          string   `json:"email" gorm:"not null"`
	Phone          string   `json:"phone"`
	Address        string   `json:"address"`
	City           string   `json:"city"`
	Country        string   `json:"country"`
	Website        string   `json:"website"`
	TaxNumber      string   `json:"tax_number"`
	RegistrationNo string   `json:"registration_no"`
	Industry       string   `json:"industry"`
	Size           string   `json:"size"` // small, medium, large, enterprise
	BaseCurrencyID uint     `json:"base_currency_id"`
	BaseCurrency   Currency `json:"base_currency" gorm:"foreignKey:BaseCurrencyID"`

	// Billing Information
	BillingPlan     string     `json:"billing_plan"`  // free, starter, professional, enterprise
	BillingCycle    string     `json:"billing_cycle"` // monthly, yearly
	SubscriptionEnd *time.Time `json:"subscription_end"`
	MaxEmployees    int        `json:"max_employees"`

	// Settings
	LogoURL      string  `json:"logo_url"`
	PayrollCycle string  `json:"payroll_cycle" gorm:"default:'monthly'"` // weekly, bi-weekly, monthly
	WorkWeekDays int     `json:"work_week_days" gorm:"default:5"`
	WorkDayHours float64 `json:"work_day_hours" gorm:"default:8"`
	OvertimeRate float64 `json:"overtime_rate" gorm:"default:1.5"`
	WeekendRate  float64 `json:"weekend_rate" gorm:"default:2.0"`

	// Status
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	IsVerified bool       `json:"is_verified" gorm:"default:false"`
	VerifiedAt *time.Time `json:"verified_at"`

	// Audit
	CreatedBy uint           `json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CompanyUser struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CompanyID uint           `json:"company_id"`
	Company   Company        `json:"company" gorm:"foreignKey:CompanyID"`
	UserID    uint           `json:"user_id"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
	Role      string         `json:"role"`       // company_admin, hr, manager, employee
	IsDefault bool           `json:"is_default"` // Default company for user login
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	JoinedAt  time.Time      `json:"joined_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// CompanySettings stores company-specific configuration
type CompanySettings struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	CompanyID uint    `json:"company_id" gorm:"unique"`
	Company   Company `json:"company" gorm:"foreignKey:CompanyID"`

	// Tax Settings
	EnablePAYE     bool   `json:"enable_paye" gorm:"default:true"`
	EnableNSSA     bool   `json:"enable_nssa" gorm:"default:true"`
	EnableAidsLevy bool   `json:"enable_aids_levy" gorm:"default:true"`
	CustomTaxRates string `json:"custom_tax_rates" gorm:"type:jsonb"` // JSON for custom tax brackets

	// Leave Settings
	LeaveYearStart     time.Time `json:"leave_year_start"`
	AllowNegativeLeave bool      `json:"allow_negative_leave" gorm:"default:false"`

	// Payroll Settings
	PayrollApprovalLevels int  `json:"payroll_approval_levels" gorm:"default:1"`
	RequireTimesheet      bool `json:"require_timesheet" gorm:"default:false"`

	// Notification Settings
	EmailNotifications bool `json:"email_notifications" gorm:"default:true"`
	SMSNotifications   bool `json:"sms_notifications" gorm:"default:false"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
