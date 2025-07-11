package handlers

import (
	"fmt"
	"gm58-hr-backend/internal/models"
	"gm58-hr-backend/internal/services/currency"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EmployeeHandler struct {
	db              *gorm.DB
	currencyService *currency.CurrencyService
}

func NewEmployeeHandler(db *gorm.DB, currencyService *currency.CurrencyService) *EmployeeHandler {
	return &EmployeeHandler{
		db:              db,
		currencyService: currencyService,
	}
}

func (eh *EmployeeHandler) GetEmployees(c *gin.Context) {
	var employees []models.Employee

	query := eh.db.Preload("Currency").Preload("Position").Preload("Department")

	if department := c.Query("department"); department != "" {
		query = query.Joins("JOIN departments ON employees.department_id = departments.id").
			Where("departments.name = ?", department)
	}

	if status := c.Query("status"); status != "" {
		query = query.Where("employment_status = ?", status)
	}

	if isActive := c.Query("is_active"); isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Employee{}).Count(&total)

	if err := query.Limit(limit).Offset(offset).Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch employees"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"employees": employees,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

func (eh *EmployeeHandler) CreateEmployee(c *gin.Context) {
	// Use a temporary struct to handle string dates and manager_id
	var tempEmployee struct {
		UserID                *uint   `json:"user_id"`
		EmployeeNumber        string  `json:"employee_number"`
		FirstName             string  `json:"first_name"`
		LastName              string  `json:"last_name"`
		MiddleName            string  `json:"middle_name"`
		NationalID            string  `json:"national_id"`
		TaxNumber             string  `json:"tax_number"`
		PassportNumber        string  `json:"passport_number"`
		Email                 string  `json:"email"`
		Phone                 string  `json:"phone"`
		AlternativePhone      string  `json:"alternative_phone"`
		Address               string  `json:"address"`
		City                  string  `json:"city"`
		Country               string  `json:"country"`
		PositionID            uint    `json:"position_id"`
		DepartmentID          uint    `json:"department_id"`
		ManagerID             string  `json:"manager_id"` // Handle as string
		BasicSalary           float64 `json:"basic_salary"`
		CurrencyID            uint    `json:"currency_id"`
		PaymentMethod         string  `json:"payment_method"`
		PaymentSchedule       string  `json:"payment_schedule"`
		BankName              string  `json:"bank_name"`
		BankAccount           string  `json:"bank_account"`
		BankBranch            string  `json:"bank_branch"`
		BankCode              string  `json:"bank_code"`
		SwiftCode             string  `json:"swift_code"`
		HireDate              string  `json:"hire_date"`          // Handle as string
		ProbationEndDate      string  `json:"probation_end_date"` // Handle as string
		ContractEndDate       string  `json:"contract_end_date"`  // Handle as string
		TerminationDate       string  `json:"termination_date"`   // Handle as string
		EmploymentType        string  `json:"employment_type"`
		EmploymentStatus      string  `json:"employment_status"`
		IsActive              bool    `json:"is_active"`
		EmergencyContactName  string  `json:"emergency_contact_name"`
		EmergencyContactPhone string  `json:"emergency_contact_phone"`
		MedicalAidNumber      string  `json:"medical_aid_number"`
		MedicalAidProvider    string  `json:"medical_aid_provider"`
	}

	if err := c.ShouldBindJSON(&tempEmployee); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to Employee struct with proper types
	employee := models.Employee{
		UserID:                tempEmployee.UserID,
		EmployeeNumber:        tempEmployee.EmployeeNumber,
		FirstName:             tempEmployee.FirstName,
		LastName:              tempEmployee.LastName,
		MiddleName:            tempEmployee.MiddleName,
		NationalID:            tempEmployee.NationalID,
		TaxNumber:             tempEmployee.TaxNumber,
		PassportNumber:        tempEmployee.PassportNumber,
		Email:                 tempEmployee.Email,
		Phone:                 tempEmployee.Phone,
		AlternativePhone:      tempEmployee.AlternativePhone,
		Address:               tempEmployee.Address,
		City:                  tempEmployee.City,
		Country:               tempEmployee.Country,
		PositionID:            tempEmployee.PositionID,
		DepartmentID:          tempEmployee.DepartmentID,
		BasicSalary:           tempEmployee.BasicSalary,
		CurrencyID:            tempEmployee.CurrencyID,
		PaymentMethod:         tempEmployee.PaymentMethod,
		PaymentSchedule:       tempEmployee.PaymentSchedule,
		BankName:              tempEmployee.BankName,
		BankAccount:           tempEmployee.BankAccount,
		BankBranch:            tempEmployee.BankBranch,
		BankCode:              tempEmployee.BankCode,
		SwiftCode:             tempEmployee.SwiftCode,
		EmploymentType:        tempEmployee.EmploymentType,
		EmploymentStatus:      tempEmployee.EmploymentStatus,
		IsActive:              tempEmployee.IsActive,
		EmergencyContactName:  tempEmployee.EmergencyContactName,
		EmergencyContactPhone: tempEmployee.EmergencyContactPhone,
		MedicalAidNumber:      tempEmployee.MedicalAidNumber,
		MedicalAidProvider:    tempEmployee.MedicalAidProvider,
	}

	// Parse ManagerID (handle empty string)
	if tempEmployee.ManagerID != "" {
		if managerID, err := strconv.ParseUint(tempEmployee.ManagerID, 10, 32); err == nil {
			id := uint(managerID)
			employee.ManagerID = &id
		}
	}

	// // Parse dates in YYYY-MM-DD format
	// if tempEmployee.HireDate != "" {
	// 	if hireDate, err := time.Parse("2006-01-02", tempEmployee.HireDate); err == nil {
	// 		employee.HireDate = hireDate
	// 	} else {
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hire_date format. Use YYYY-MM-DD"})
	// 		return
	// 	}
	// }

	// if tempEmployee.ProbationEndDate != "" {
	// 	if probationEndDate, err := time.Parse("2006-01-02", tempEmployee.ProbationEndDate); err == nil {
	// 		employee.ProbationEndDate = &probationEndDate
	// 	}
	// }

	// if tempEmployee.ContractEndDate != "" {
	// 	if contractEndDate, err := time.Parse("2006-01-02", tempEmployee.ContractEndDate); err == nil {
	// 		employee.ContractEndDate = &contractEndDate
	// 	}
	// }

	// if tempEmployee.TerminationDate != "" {
	// 	if terminationDate, err := time.Parse("2006-01-02", tempEmployee.TerminationDate); err == nil {
	// 		employee.TerminationDate = &terminationDate
	// 	}
	// }

	// Validate currency
	var currency models.Currency
	if err := eh.db.First(&currency, employee.CurrencyID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid currency"})
		return
	}

	// Generate employee number if not provided
	if employee.EmployeeNumber == "" {
		employee.EmployeeNumber = eh.generateEmployeeNumber()
	}

	// Create employee
	if err := eh.db.Create(&employee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create employee"})
		return
	}

	// Load relationships
	eh.db.Preload("Currency").Preload("Position").Preload("Department").First(&employee, employee.ID)

	c.JSON(http.StatusCreated, employee)
}

func (eh *EmployeeHandler) GetEmployee(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	var employee models.Employee
	if err := eh.db.Preload("Currency").Preload("Position").Preload("Department").
		Preload("Manager").First(&employee, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch employee"})
		return
	}

	c.JSON(http.StatusOK, employee)
}

func (eh *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	var employee models.Employee
	if err := eh.db.First(&employee, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch employee"})
		return
	}

	// Use the same temporary struct for updates
	var tempEmployee struct {
		UserID                *uint   `json:"user_id"`
		EmployeeNumber        string  `json:"employee_number"`
		FirstName             string  `json:"first_name"`
		LastName              string  `json:"last_name"`
		MiddleName            string  `json:"middle_name"`
		NationalID            string  `json:"national_id"`
		TaxNumber             string  `json:"tax_number"`
		PassportNumber        string  `json:"passport_number"`
		Email                 string  `json:"email"`
		Phone                 string  `json:"phone"`
		AlternativePhone      string  `json:"alternative_phone"`
		Address               string  `json:"address"`
		City                  string  `json:"city"`
		Country               string  `json:"country"`
		PositionID            uint    `json:"position_id"`
		DepartmentID          uint    `json:"department_id"`
		ManagerID             string  `json:"manager_id"`
		BasicSalary           float64 `json:"basic_salary"`
		CurrencyID            uint    `json:"currency_id"`
		PaymentMethod         string  `json:"payment_method"`
		PaymentSchedule       string  `json:"payment_schedule"`
		BankName              string  `json:"bank_name"`
		BankAccount           string  `json:"bank_account"`
		BankBranch            string  `json:"bank_branch"`
		BankCode              string  `json:"bank_code"`
		SwiftCode             string  `json:"swift_code"`
		HireDate              string  `json:"hire_date"`
		ProbationEndDate      string  `json:"probation_end_date"`
		ContractEndDate       string  `json:"contract_end_date"`
		TerminationDate       string  `json:"termination_date"`
		EmploymentType        string  `json:"employment_type"`
		EmploymentStatus      string  `json:"employment_status"`
		IsActive              bool    `json:"is_active"`
		EmergencyContactName  string  `json:"emergency_contact_name"`
		EmergencyContactPhone string  `json:"emergency_contact_phone"`
		MedicalAidNumber      string  `json:"medical_aid_number"`
		MedicalAidProvider    string  `json:"medical_aid_provider"`
	}

	if err := c.ShouldBindJSON(&tempEmployee); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields selectively (only if provided and not empty)
	if tempEmployee.FirstName != "" {
		employee.FirstName = tempEmployee.FirstName
	}
	if tempEmployee.LastName != "" {
		employee.LastName = tempEmployee.LastName
	}
	employee.MiddleName = tempEmployee.MiddleName // Allow empty

	if tempEmployee.Email != "" {
		employee.Email = tempEmployee.Email
	}
	if tempEmployee.Phone != "" {
		employee.Phone = tempEmployee.Phone
	}
	employee.AlternativePhone = tempEmployee.AlternativePhone // Allow empty
	employee.Address = tempEmployee.Address                   // Allow empty

	if tempEmployee.City != "" {
		employee.City = tempEmployee.City
	}
	if tempEmployee.Country != "" {
		employee.Country = tempEmployee.Country
	}
	if tempEmployee.BasicSalary > 0 {
		employee.BasicSalary = tempEmployee.BasicSalary
	}
	if tempEmployee.EmploymentType != "" {
		employee.EmploymentType = tempEmployee.EmploymentType
	}
	if tempEmployee.EmploymentStatus != "" {
		employee.EmploymentStatus = tempEmployee.EmploymentStatus
	}

	// Handle ManagerID
	if tempEmployee.ManagerID != "" {
		if managerID, err := strconv.ParseUint(tempEmployee.ManagerID, 10, 32); err == nil {
			id := uint(managerID)
			employee.ManagerID = &id
		}
	} else {
		employee.ManagerID = nil
	}

	// // Handle date updates
	// if tempEmployee.HireDate != "" {
	// 	if hireDate, err := time.Parse("2006-01-02", tempEmployee.HireDate); err == nil {
	// 		employee.HireDate = hireDate
	// 	}
	// }

	// if tempEmployee.ProbationEndDate != "" {
	// 	if probationEndDate, err := time.Parse("2006-01-02", tempEmployee.ProbationEndDate); err == nil {
	// 		employee.ProbationEndDate = &probationEndDate
	// 	}
	// }

	// if tempEmployee.ContractEndDate != "" {
	// 	if contractEndDate, err := time.Parse("2006-01-02", tempEmployee.ContractEndDate); err == nil {
	// 		employee.ContractEndDate = &contractEndDate
	// 	}
	// }

	// if tempEmployee.TerminationDate != "" {
	// 	if terminationDate, err := time.Parse("2006-01-02", tempEmployee.TerminationDate); err == nil {
	// 		employee.TerminationDate = &terminationDate
	// 	}
	// }

	// Validate currency if changed
	if tempEmployee.CurrencyID != 0 && tempEmployee.CurrencyID != employee.CurrencyID {
		var currency models.Currency
		if err := eh.db.First(&currency, tempEmployee.CurrencyID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid currency"})
			return
		}
		employee.CurrencyID = tempEmployee.CurrencyID
	}

	// Save updates
	if err := eh.db.Save(&employee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update employee"})
		return
	}

	// Load relationships
	eh.db.Preload("Currency").Preload("Position").Preload("Department").First(&employee, employee.ID)

	c.JSON(http.StatusOK, employee)
}

func (eh *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	var employee models.Employee
	if err := eh.db.First(&employee, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch employee"})
		return
	}

	if err := eh.db.Delete(&employee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete employee"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Employee deleted successfully"})
}

func (eh *EmployeeHandler) generateEmployeeNumber() string {
	var count int64
	eh.db.Model(&models.Employee{}).Count(&count)
	return fmt.Sprintf("EMP%06d", count+1)
}

func (eh *EmployeeHandler) GetEmployeePayslips(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	var payslips []models.Payslip
	query := eh.db.Preload("PayrollPeriod").Preload("Currency").Where("employee_id = ?", uint(id))

	if year := c.Query("year"); year != "" {
		query = query.Joins("JOIN payroll_periods ON payslips.payroll_period_id = payroll_periods.id").
			Where("payroll_periods.year = ?", year)
	}

	if err := query.Order("payroll_periods.year DESC, payroll_periods.month DESC").Find(&payslips).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payslips"})
		return
	}

	c.JSON(http.StatusOK, payslips)
}
