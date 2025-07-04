package handlers

import (
	"gm58-hr-backend/internal/api/middleware"
	"gm58-hr-backend/internal/models"
	"gm58-hr-backend/internal/services/payroll"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PayrollHandler struct {
	db        *gorm.DB
	processor *payroll.PayrollProcessor
}

func NewPayrollHandler(db *gorm.DB, processor *payroll.PayrollProcessor) *PayrollHandler {
	return &PayrollHandler{
		db:        db,
		processor: processor,
	}
}

// Key updates to payroll handler for multi-company support

func (ph *PayrollHandler) CreatePeriod(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)

	var period models.PayrollPeriod
	if err := c.ShouldBindJSON(&period); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set company ID
	period.CompanyID = companyID

	// Check if period already exists for this company
	var existingPeriod models.PayrollPeriod
	if err := ph.db.Where("company_id = ? AND year = ? AND month = ?",
		companyID, period.Year, period.Month).First(&existingPeriod).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Payroll period already exists for this month"})
		return
	}

	period.StartDate = time.Date(period.Year, time.Month(period.Month), 1, 0, 0, 0, 0, time.UTC)
	period.EndDate = period.StartDate.AddDate(0, 1, -1)

	if err := ph.db.Create(&period).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payroll period"})
		return
	}

	c.JSON(http.StatusCreated, period)
}

func (ph *PayrollHandler) GetPeriods(c *gin.Context) {
	// companyID := middleware.GetCompanyID(c)

	var periods []models.PayrollPeriod

	query := ph.db.Scopes(middleware.CompanyScope(c))

	if year := c.Query("year"); year != "" {
		query = query.Where("year = ?", year)
	}

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("year DESC, month DESC").Find(&periods).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch periods"})
		return
	}

	c.JSON(http.StatusOK, periods)
}

func (ph *PayrollHandler) ProcessPayroll(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	periodID, err := strconv.ParseUint(c.Param("periodId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}

	// Verify period belongs to company
	var period models.PayrollPeriod
	if err := ph.db.Where("id = ? AND company_id = ?", periodID, companyID).First(&period).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payroll period not found"})
		return
	}

	if err := ph.processor.ProcessPayrollForCompany(uint(periodID), companyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payroll processed successfully"})
}

func (ph *PayrollHandler) GetPayslips(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	periodID, err := strconv.ParseUint(c.Param("periodId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}

	// Verify period belongs to company
	var period models.PayrollPeriod
	if err := ph.db.Where("id = ? AND company_id = ?", periodID, companyID).First(&period).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payroll period not found"})
		return
	}

	var payslips []models.Payslip
	if err := ph.db.Preload("Employee").Preload("Currency").Preload("PayrollPeriod").
		Where("payroll_period_id = ? AND company_id = ?", periodID, companyID).
		Find(&payslips).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payslips"})
		return
	}

	c.JSON(http.StatusOK, payslips)
}

func (ph *PayrollHandler) GetPayrollSummary(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	periodID, err := strconv.ParseUint(c.Param("periodId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}

	// Verify period belongs to company
	var period models.PayrollPeriod
	if err := ph.db.Where("id = ? AND company_id = ?", periodID, companyID).First(&period).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payroll period not found"})
		return
	}

	summary, err := ph.processor.GetPayrollSummaryForCompany(uint(periodID), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payroll summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (ph *PayrollHandler) ApprovePayroll(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	periodID, err := strconv.ParseUint(c.Param("periodId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
		return
	}

	// Verify period belongs to company
	var period models.PayrollPeriod
	if err := ph.db.Where("id = ? AND company_id = ?", periodID, companyID).First(&period).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payroll period not found"})
		return
	}

	userID := c.GetUint("user_id")
	if err := ph.processor.ApprovePayroll(uint(periodID), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payroll approved successfully"})
}

func (ph *PayrollHandler) GetPayslip(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	payslipID, err := strconv.ParseUint(c.Param("payslipId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payslip ID"})
		return
	}

	var payslip models.Payslip
	if err := ph.db.Preload("Employee").Preload("Currency").Preload("PayrollPeriod").
		Where("id = ? AND company_id = ?", uint(payslipID), companyID).
		First(&payslip).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payslip not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payslip"})
		return
	}

	// Check if user has permission to view this payslip
	userID := c.GetUint("user_id")
	companyRole := middleware.GetCompanyRole(c)

	// Check if user is viewing their own payslip
	var employee models.Employee
	if err := ph.db.Where("id = ? AND user_id = ?", payslip.EmployeeID, userID).First(&employee).Error; err != nil {
		// Not their payslip, check if they have permission
		if companyRole != "company_admin" && companyRole != "hr" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this payslip"})
			return
		}
	}

	c.JSON(http.StatusOK, payslip)
}

// func (ph *PayrollHandler) CreatePeriod(c *gin.Context) {
// 	var period models.PayrollPeriod
// 	if err := c.ShouldBindJSON(&period); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	var existingPeriod models.PayrollPeriod
// 	if err := ph.db.Where("year = ? AND month = ?", period.Year, period.Month).First(&existingPeriod).Error; err == nil {
// 		c.JSON(http.StatusConflict, gin.H{"error": "Payroll period already exists"})
// 		return
// 	}

// 	period.StartDate = time.Date(period.Year, time.Month(period.Month), 1, 0, 0, 0, 0, time.UTC)
// 	period.EndDate = period.StartDate.AddDate(0, 1, -1)

// 	if err := ph.db.Create(&period).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payroll period"})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, period)
// }

// func (ph *PayrollHandler) GetPeriods(c *gin.Context) {
// 	var periods []models.PayrollPeriod

// 	query := ph.db.Model(&models.PayrollPeriod{})

// 	if year := c.Query("year"); year != "" {
// 		query = query.Where("year = ?", year)
// 	}

// 	if status := c.Query("status"); status != "" {
// 		query = query.Where("status = ?", status)
// 	}

// 	if err := query.Order("year DESC, month DESC").Find(&periods).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch periods"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, periods)
// }

// func (ph *PayrollHandler) ProcessPayroll(c *gin.Context) {
// 	periodID, err := strconv.ParseUint(c.Param("periodId"), 10, 32)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
// 		return
// 	}

// 	if err := ph.processor.ProcessPayroll(uint(periodID)); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Payroll processed successfully"})
// }

// func (ph *PayrollHandler) GetPayslips(c *gin.Context) {
// 	periodID, err := strconv.ParseUint(c.Param("periodId"), 10, 32)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
// 		return
// 	}

// 	var payslips []models.Payslip
// 	if err := ph.db.Preload("Employee").Preload("Currency").Preload("PayrollPeriod").
// 		Where("payroll_period_id = ?", periodID).Find(&payslips).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payslips"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, payslips)
// }

// func (ph *PayrollHandler) GetPayrollSummary(c *gin.Context) {
// 	periodID, err := strconv.ParseUint(c.Param("periodId"), 10, 32)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
// 		return
// 	}

// 	summary, err := ph.processor.GetPayrollSummary(uint(periodID))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payroll summary"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, summary)
// }

// func (ph *PayrollHandler) ApprovePayroll(c *gin.Context) {
// 	periodID, err := strconv.ParseUint(c.Param("periodId"), 10, 32)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period ID"})
// 		return
// 	}

// 	userID := c.GetUint("user_id")
// 	if err := ph.processor.ApprovePayroll(uint(periodID), userID); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Payroll approved successfully"})
// }

// func (ph *PayrollHandler) GetPayslip(c *gin.Context) {
// 	payslipID, err := strconv.ParseUint(c.Param("payslipId"), 10, 32)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payslip ID"})
// 		return
// 	}

// 	var payslip models.Payslip
// 	if err := ph.db.Preload("Employee").Preload("Currency").Preload("PayrollPeriod").
// 		First(&payslip, uint(payslipID)).Error; err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Payslip not found"})
// 			return
// 		}
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payslip"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, payslip)
// }
