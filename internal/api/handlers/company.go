package handlers

import (
	"gm58-hr-backend/internal/api/middleware"
	"gm58-hr-backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CompanyHandler struct {
	db *gorm.DB
}

func NewCompanyHandler(db *gorm.DB) *CompanyHandler {
	return &CompanyHandler{
		db: db,
	}
}

// GetCompany returns the current company details
func (ch *CompanyHandler) GetCompany(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)

	var company models.Company
	if err := ch.db.Preload("BaseCurrency").First(&company, companyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	c.JSON(http.StatusOK, company)
}

// UpdateCompany updates company details
func (ch *CompanyHandler) UpdateCompany(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)

	var company models.Company
	if err := ch.db.First(&company, companyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	var updateData models.Company
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prevent updating sensitive fields
	updateData.ID = company.ID
	updateData.Code = company.Code
	updateData.CreatedBy = company.CreatedBy

	if err := ch.db.Model(&company).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
		return
	}

	ch.db.Preload("BaseCurrency").First(&company, company.ID)
	c.JSON(http.StatusOK, company)
}

// GetCompanySettings returns company-specific settings
func (ch *CompanyHandler) GetCompanySettings(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)

	var settings models.CompanySettings
	if err := ch.db.Where("company_id = ?", companyID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default settings
			settings = models.CompanySettings{
				CompanyID:             companyID,
				EnablePAYE:            true,
				EnableNSSA:            true,
				EnableAidsLevy:        true,
				PayrollApprovalLevels: 1,
				EmailNotifications:    true,
			}
			ch.db.Create(&settings)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
			return
		}
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateCompanySettings updates company-specific settings
func (ch *CompanyHandler) UpdateCompanySettings(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)

	var settings models.CompanySettings
	if err := ch.db.Where("company_id = ?", companyID).First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			settings.CompanyID = companyID
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
			return
		}
	}

	var updateData models.CompanySettings
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateData.ID = settings.ID
	updateData.CompanyID = companyID

	if settings.ID == 0 {
		if err := ch.db.Create(&updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create settings"})
			return
		}
	} else {
		if err := ch.db.Model(&settings).Updates(updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
			return
		}
	}

	c.JSON(http.StatusOK, updateData)
}

// GetCompanyUsers returns all users in the company
func (ch *CompanyHandler) GetCompanyUsers(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)

	var companyUsers []models.CompanyUser
	query := ch.db.Preload("User").Where("company_id = ?", companyID)

	// Filter by role
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	// Filter by active status
	if isActive := c.Query("is_active"); isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	var total int64
	query.Count(&total)

	if err := query.Limit(limit).Offset(offset).Find(&companyUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	users := make([]gin.H, len(companyUsers))
	for i, cu := range companyUsers {
		users[i] = gin.H{
			"id":         cu.ID,
			"user_id":    cu.UserID,
			"username":   cu.User.Username,
			"email":      cu.User.Email,
			"role":       cu.Role,
			"is_active":  cu.IsActive,
			"is_default": cu.IsDefault,
			"joined_at":  cu.JoinedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// AddUserToCompany adds an existing user to the company
func (ch *CompanyHandler) AddUserToCompany(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)

	var req struct {
		Email string `json:"email" binding:"required,email"`
		Role  string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := ch.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user already in company
	var existingCU models.CompanyUser
	if err := ch.db.Where("company_id = ? AND user_id = ?", companyID, user.ID).First(&existingCU).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists in company"})
		return
	}

	// Add user to company
	companyUser := models.CompanyUser{
		CompanyID: companyID,
		UserID:    user.ID,
		Role:      req.Role,
		IsActive:  true,
	}

	if err := ch.db.Create(&companyUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to company"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User added to company successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     req.Role,
		},
	})
}

// UpdateUserRole updates a user's role in the company
func (ch *CompanyHandler) UpdateUserRole(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var companyUser models.CompanyUser
	if err := ch.db.Where("company_id = ? AND user_id = ?", companyID, uint(userID)).First(&companyUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in company"})
		return
	}

	companyUser.Role = req.Role
	if err := ch.db.Save(&companyUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

// RemoveUserFromCompany removes a user from the company
func (ch *CompanyHandler) RemoveUserFromCompany(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Don't allow removing the last admin
	var adminCount int64
	ch.db.Model(&models.CompanyUser{}).
		Where("company_id = ? AND role = 'company_admin' AND is_active = ?", companyID, true).
		Count(&adminCount)

	var companyUser models.CompanyUser
	if err := ch.db.Where("company_id = ? AND user_id = ?", companyID, uint(userID)).First(&companyUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in company"})
		return
	}

	if companyUser.Role == "company_admin" && adminCount <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove the last admin"})
		return
	}

	if err := ch.db.Delete(&companyUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user from company"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed from company successfully"})
}

// GetCompanyStats returns company statistics
func (ch *CompanyHandler) GetCompanyStats(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)

	stats := gin.H{}

	// Employee count
	var employeeCount int64
	ch.db.Model(&models.Employee{}).Where("company_id = ? AND is_active = ?", companyID, true).Count(&employeeCount)
	stats["total_employees"] = employeeCount

	// Department count
	var departmentCount int64
	ch.db.Model(&models.Department{}).Where("company_id = ? AND is_active = ?", companyID, true).Count(&departmentCount)
	stats["total_departments"] = departmentCount

	// Position count
	var positionCount int64
	ch.db.Model(&models.Position{}).Where("company_id = ? AND is_active = ?", companyID, true).Count(&positionCount)
	stats["total_positions"] = positionCount

	// User count
	var userCount int64
	ch.db.Model(&models.CompanyUser{}).Where("company_id = ? AND is_active = ?", companyID, true).Count(&userCount)
	stats["total_users"] = userCount

	// Recent payroll
	var lastPayroll models.PayrollPeriod
	ch.db.Where("company_id = ?", companyID).Order("year DESC, month DESC").First(&lastPayroll)
	if lastPayroll.ID != 0 {
		stats["last_payroll"] = gin.H{
			"year":   lastPayroll.Year,
			"month":  lastPayroll.Month,
			"status": lastPayroll.Status,
		}
	}

	// Company details
	var company models.Company
	ch.db.First(&company, companyID)
	stats["company"] = gin.H{
		"name":           company.Name,
		"billing_plan":   company.BillingPlan,
		"max_employees":  company.MaxEmployees,
		"employee_usage": employeeCount,
	}

	c.JSON(http.StatusOK, stats)
}

// For super admin only - list all companies
func (ch *CompanyHandler) ListCompanies(c *gin.Context) {
	if !middleware.IsSuperAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Super admin access required"})
		return
	}

	var companies []models.Company
	query := ch.db.Preload("BaseCurrency")

	// Filter by active status
	if isActive := c.Query("is_active"); isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	// Filter by billing plan
	if plan := c.Query("billing_plan"); plan != "" {
		query = query.Where("billing_plan = ?", plan)
	}

	// Search by name or code
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Company{}).Count(&total)

	if err := query.Limit(limit).Offset(offset).Find(&companies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch companies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"companies": companies,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}
