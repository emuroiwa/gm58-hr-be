package models

import (
	"time"
	"gorm.io/gorm"
)

type LeaveType struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	Name          string `json:"name" gorm:"unique;not null"`
	Description   string `json:"description"`
	DaysPerYear   int    `json:"days_per_year"`
	IsPaid        bool   `json:"is_paid" gorm:"default:true"`
	CarryForward  bool   `json:"carry_forward" gorm:"default:false"`
	MaxCarryDays  int    `json:"max_carry_days"`
	RequiresApproval bool `json:"requires_approval" gorm:"default:true"`
	IsActive      bool   `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

type LeaveRequest struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	EmployeeID    uint      `json:"employee_id"`
	Employee      Employee  `json:"employee" gorm:"foreignKey:EmployeeID"`
	LeaveTypeID   uint      `json:"leave_type_id"`
	LeaveType     LeaveType `json:"leave_type" gorm:"foreignKey:LeaveTypeID"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	DaysRequested int       `json:"days_requested"`
	Reason        string    `json:"reason"`
	Status        string    `json:"status" gorm:"default:'pending'"` // pending, approved, rejected, cancelled
	ApprovedBy    *uint     `json:"approved_by"`
	Approver      *Employee `json:"approver,omitempty" gorm:"foreignKey:ApprovedBy"`
	ApprovedAt    *time.Time `json:"approved_at"`
	RejectionReason string  `json:"rejection_reason"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

type TaxCertificate struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	EmployeeID        uint      `json:"employee_id"`
	Employee          Employee  `json:"employee" gorm:"foreignKey:EmployeeID"`
	Year              int       `json:"year"`
	TotalEarnings     float64   `json:"total_earnings" gorm:"type:decimal(15,2)"`
	TotalTax          float64   `json:"total_tax" gorm:"type:decimal(15,2)"`
	CurrencyID        uint      `json:"currency_id"`
	Currency          Currency  `json:"currency" gorm:"foreignKey:CurrencyID"`
	CertificateNumber string    `json:"certificate_number"`
	IssuedAt          time.Time `json:"issued_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type AuditLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      *uint     `json:"user_id"`
	User        *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Action      string    `json:"action"` // CREATE, UPDATE, DELETE, LOGIN, LOGOUT
	EntityType  string    `json:"entity_type"` // Employee, Payslip, etc.
	EntityID    *uint     `json:"entity_id"`
	OldValues   string    `json:"old_values" gorm:"type:jsonb"`
	NewValues   string    `json:"new_values" gorm:"type:jsonb"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	CreatedAt   time.Time `json:"created_at"`
}
