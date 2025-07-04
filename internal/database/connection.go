package database

import (
	"gm58-hr-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// Company models
		&models.Company{},
		&models.CompanyUser{},
		&models.CompanySettings{},

		// User and Employee models
		&models.User{},
		&models.Employee{},
		&models.Department{},
		&models.Position{},

		// Currency models
		&models.Currency{},
		&models.ExchangeRate{},

		// Payroll models
		&models.PayrollPeriod{},
		&models.Payslip{},
		&models.Allowance{},
		&models.Deduction{},

		// Leave models
		&models.LeaveType{},
		&models.LeaveRequest{},

		// Other models
		&models.TaxCertificate{},
		&models.AuditLog{},
	)
}
