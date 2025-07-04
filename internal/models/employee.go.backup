package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"unique;not null"`
	Email     string         `json:"email" gorm:"unique;not null"`
	Password  string         `json:"-" gorm:"not null"`
	Role      string         `json:"role" gorm:"not null;default:'employee'"` // admin, hr, employee
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Department struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"unique;not null"`
	Description string         `json:"description"`
	ManagerID   *uint          `json:"manager_id"`
	Manager     *Employee      `json:"manager,omitempty" gorm:"foreignKey:ManagerID"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type Position struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Title        string         `json:"title" gorm:"not null"`
	DepartmentID uint           `json:"department_id"`
	Department   Department     `json:"department" gorm:"foreignKey:DepartmentID"`
	Description  string         `json:"description"`
	MinSalary    float64        `json:"min_salary" gorm:"type:decimal(15,2)"`
	MaxSalary    float64        `json:"max_salary" gorm:"type:decimal(15,2)"`
	CurrencyID   uint           `json:"currency_id"`
	Currency     Currency       `json:"currency" gorm:"foreignKey:CurrencyID"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type Employee struct {
	ID               uint   `json:"id" gorm:"primaryKey"`
	UserID           *uint  `json:"user_id"`
	User             *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	EmployeeNumber   string `json:"employee_number" gorm:"unique;not null"`
	FirstName        string `json:"first_name" gorm:"not null"`
	LastName         string `json:"last_name" gorm:"not null"`
	MiddleName       string `json:"middle_name"`
	NationalID       string `json:"national_id" gorm:"unique"`
	TaxNumber        string `json:"tax_number"`
	PassportNumber   string `json:"passport_number"`
	Email            string `json:"email" gorm:"unique"`
	Phone            string `json:"phone"`
	AlternativePhone string `json:"alternative_phone"`
	Address          string `json:"address"`
	City             string `json:"city"`
	Country          string `json:"country" gorm:"default:'Local'"`

	// Employment Details
	PositionID   uint       `json:"position_id"`
	Position     Position   `json:"position" gorm:"foreignKey:PositionID"`
	DepartmentID uint       `json:"department_id"`
	Department   Department `json:"department" gorm:"foreignKey:DepartmentID"`
	ManagerID    *uint      `json:"manager_id"`
	Manager      *Employee  `json:"manager,omitempty" gorm:"foreignKey:ManagerID"`

	// Salary Information
	BasicSalary     float64  `json:"basic_salary" gorm:"type:decimal(15,2)"`
	CurrencyID      uint     `json:"currency_id"`
	Currency        Currency `json:"currency" gorm:"foreignKey:CurrencyID"`
	PaymentMethod   string   `json:"payment_method" gorm:"default:'bank_transfer'"`
	PaymentSchedule string   `json:"payment_schedule" gorm:"default:'monthly'"` // weekly, bi-weekly, monthly

	// Bank Details
	BankName    string `json:"bank_name"`
	BankAccount string `json:"bank_account"`
	BankBranch  string `json:"bank_branch"`
	BankCode    string `json:"bank_code"`
	SwiftCode   string `json:"swift_code"`

	// Employment Dates
	HireDate         string `json:"hire_date"`
	ProbationEndDate string `json:"probation_end_date"`
	ContractEndDate  string `json:"contract_end_date"`
	TerminationDate  string `json:"termination_date"`

	// Status
	EmploymentType   string `json:"employment_type" gorm:"default:'permanent'"` // permanent, contract, temporary
	EmploymentStatus string `json:"employment_status" gorm:"default:'active'"`  // active, suspended, terminated
	IsActive         bool   `json:"is_active" gorm:"default:true"`

	// Medical and Emergency
	EmergencyContactName  string `json:"emergency_contact_name"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
	MedicalAidNumber      string `json:"medical_aid_number"`
	MedicalAidProvider    string `json:"medical_aid_provider"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Payslips        []Payslip        `json:"payslips,omitempty" gorm:"foreignKey:EmployeeID"`
	Allowances      []Allowance      `json:"allowances,omitempty" gorm:"foreignKey:EmployeeID"`
	Deductions      []Deduction      `json:"deductions,omitempty" gorm:"foreignKey:EmployeeID"`
	LeaveRequests   []LeaveRequest   `json:"leave_requests,omitempty" gorm:"foreignKey:EmployeeID"`
	TaxCertificates []TaxCertificate `json:"tax_certificates,omitempty" gorm:"foreignKey:EmployeeID"`
}

func (e *Employee) FullName() string {
	if e.MiddleName != "" {
		return e.FirstName + " " + e.MiddleName + " " + e.LastName
	}
	return e.FirstName + " " + e.LastName
}
