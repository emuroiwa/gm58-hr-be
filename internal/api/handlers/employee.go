package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"gm58-hr-backend/internal/models"
	"gm58-hr-backend/internal/services/currency"

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
	var employee models.Employee
	if err := c.ShouldBindJSON(&employee); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var currency models.Currency
	if err := eh.db.First(&currency, employee.CurrencyID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid currency"})
		return
	}

	if employee.EmployeeNumber == "" {
		employee.EmployeeNumber = eh.generateEmployeeNumber()
	}

	if err := eh.db.Create(&employee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create employee"})
		return
	}

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

	var updateData models.Employee
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updateData.CurrencyID != 0 && updateData.CurrencyID != employee.CurrencyID {
		var currency models.Currency
		if err := eh.db.First(&currency, updateData.CurrencyID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid currency"})
			return
		}
	}

	if err := eh.db.Model(&employee).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update employee"})
		return
	}

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
