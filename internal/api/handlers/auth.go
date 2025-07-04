package handlers

import (
	"gm58-hr-backend/internal/api/middleware"
	"gm58-hr-backend/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
}

type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	CompanyID *uint  `json:"company_id"` // Optional: specific company to login to
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

type CompanyRegistrationRequest struct {
	// Company details
	CompanyName  string `json:"company_name" binding:"required"`
	CompanyCode  string `json:"company_code" binding:"required"`
	CompanyEmail string `json:"company_email" binding:"required,email"`
	CompanyPhone string `json:"company_phone"`
	Industry     string `json:"industry"`
	Size         string `json:"size"`

	// Admin user details
	AdminUsername  string `json:"admin_username" binding:"required"`
	AdminEmail     string `json:"admin_email" binding:"required,email"`
	AdminPassword  string `json:"admin_password" binding:"required,min=6"`
	AdminFirstName string `json:"admin_first_name" binding:"required"`
	AdminLastName  string `json:"admin_last_name" binding:"required"`
}

func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (ah *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := ah.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is deactivated"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Get user's companies
	var companyUsers []models.CompanyUser
	ah.db.Preload("Company").Where("user_id = ? AND is_active = ?", user.ID, true).Find(&companyUsers)

	if len(companyUsers) == 0 && user.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "No active company membership found"})
		return
	}

	// Determine which company to use
	var selectedCompany *models.CompanyUser
	if req.CompanyID != nil {
		// User specified a company
		for _, cu := range companyUsers {
			if cu.CompanyID == *req.CompanyID {
				selectedCompany = &cu
				break
			}
		}
		if selectedCompany == nil && user.Role != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to specified company"})
			return
		}
	} else if len(companyUsers) == 1 {
		// User has only one company
		selectedCompany = &companyUsers[0]
	} else {
		// User has multiple companies, return list for selection
		companies := make([]gin.H, len(companyUsers))
		for i, cu := range companyUsers {
			companies[i] = gin.H{
				"id":   cu.Company.ID,
				"name": cu.Company.Name,
				"code": cu.Company.Code,
				"role": cu.Role,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"requires_company_selection": true,
			"companies":                  companies,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
		})
		return
	}

	now := time.Now()
	user.LastLogin = &now
	ah.db.Save(&user)

	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role, ah.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	}

	if selectedCompany != nil {
		response["company"] = gin.H{
			"id":   selectedCompany.Company.ID,
			"name": selectedCompany.Company.Name,
			"code": selectedCompany.Company.Code,
			"role": selectedCompany.Role,
		}
	}

	c.JSON(http.StatusOK, response)
}

func (ah *AuthHandler) RegisterCompany(c *gin.Context) {
	var req CompanyRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if company code already exists
	var existingCompany models.Company
	if err := ah.db.Where("code = ?", req.CompanyCode).First(&existingCompany).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Company code already exists"})
		return
	}

	// Check if admin user already exists
	var existingUser models.User
	if err := ah.db.Where("username = ? OR email = ?", req.AdminUsername, req.AdminEmail).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Start transaction
	tx := ah.db.Begin()

	// Create company
	var baseCurrency models.Currency
	tx.Where("is_base_currency = ?", true).First(&baseCurrency)

	company := models.Company{
		Name:           req.CompanyName,
		Code:           req.CompanyCode,
		Email:          req.CompanyEmail,
		Phone:          req.CompanyPhone,
		Industry:       req.Industry,
		Size:           req.Size,
		BillingPlan:    "free",
		MaxEmployees:   10,
		BaseCurrencyID: baseCurrency.ID,
		IsActive:       true,
	}

	if err := tx.Create(&company).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company"})
		return
	}

	// Create admin user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	adminUser := models.User{
		Username: req.AdminUsername,
		Email:    req.AdminEmail,
		Password: string(hashedPassword),
		Role:     "company_admin",
		IsActive: true,
	}

	if err := tx.Create(&adminUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user"})
		return
	}

	// Create company-user relationship
	companyUser := models.CompanyUser{
		CompanyID: company.ID,
		UserID:    adminUser.ID,
		Role:      "company_admin",
		IsDefault: true,
		IsActive:  true,
		JoinedAt:  time.Now(),
	}

	if err := tx.Create(&companyUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company-user relationship"})
		return
	}

	// Create admin employee record
	employee := models.Employee{
		CompanyID:      company.ID,
		UserID:         &adminUser.ID,
		EmployeeNumber: "EMP001",
		FirstName:      req.AdminFirstName,
		LastName:       req.AdminLastName,
		Email:          req.AdminEmail,
		CurrencyID:     baseCurrency.ID,
		IsActive:       true,
		HireDate:       time.Now().Format("2006-01-02"),
	}

	if err := tx.Create(&employee).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create employee record"})
		return
	}

	// Create default departments
	departments := []models.Department{
		{CompanyID: company.ID, Name: "Management", Description: "Executive management", IsActive: true},
		{CompanyID: company.ID, Name: "Operations", Description: "Day-to-day operations", IsActive: true},
		{CompanyID: company.ID, Name: "Finance", Description: "Financial management", IsActive: true},
	}

	for _, dept := range departments {
		if err := tx.Create(&dept).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create departments"})
			return
		}
	}

	// Create default leave types
	leaveTypes := []models.LeaveType{
		{CompanyID: company.ID, Name: "Annual Leave", DaysPerYear: 21, IsPaid: true, IsActive: true},
		{CompanyID: company.ID, Name: "Sick Leave", DaysPerYear: 10, IsPaid: true, IsActive: true},
		{CompanyID: company.ID, Name: "Personal Leave", DaysPerYear: 3, IsPaid: false, IsActive: true},
	}

	for _, lt := range leaveTypes {
		if err := tx.Create(&lt).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create leave types"})
			return
		}
	}

	// Commit transaction
	tx.Commit()

	// Generate token
	token, _ := middleware.GenerateToken(adminUser.ID, adminUser.Username, adminUser.Role, ah.jwtSecret)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Company registered successfully",
		"token":   token,
		"company": gin.H{
			"id":   company.ID,
			"name": company.Name,
			"code": company.Code,
		},
		"user": gin.H{
			"id":       adminUser.ID,
			"username": adminUser.Username,
			"email":    adminUser.Email,
			"role":     adminUser.Role,
		},
	})
}

func (ah *AuthHandler) GetUserCompanies(c *gin.Context) {
	userID := c.GetUint("user_id")

	var companyUsers []models.CompanyUser
	ah.db.Preload("Company").Where("user_id = ? AND is_active = ?", userID, true).Find(&companyUsers)

	companies := make([]gin.H, len(companyUsers))
	for i, cu := range companyUsers {
		companies[i] = gin.H{
			"id":         cu.Company.ID,
			"name":       cu.Company.Name,
			"code":       cu.Company.Code,
			"role":       cu.Role,
			"is_default": cu.IsDefault,
			"joined_at":  cu.JoinedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"companies": companies})
}

func (ah *AuthHandler) SwitchCompany(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		CompanyID uint `json:"company_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify user has access to the company
	var companyUser models.CompanyUser
	if err := ah.db.Preload("Company").Where("user_id = ? AND company_id = ? AND is_active = ?",
		userID, req.CompanyID, true).First(&companyUser).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this company"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Company switched successfully",
		"company": gin.H{
			"id":   companyUser.Company.ID,
			"name": companyUser.Company.Name,
			"code": companyUser.Company.Code,
			"role": companyUser.Role,
		},
	})
}

func (ah *AuthHandler) Register(c *gin.Context) {
	// This endpoint now requires company context
	companyID := middleware.GetCompanyID(c)
	companyRole := middleware.GetCompanyRole(c)

	if companyRole != "company_admin" && !middleware.IsSuperAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only company admins can register new users"})
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	if err := ah.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		// User exists, check if already in this company
		var companyUser models.CompanyUser
		if err := ah.db.Where("user_id = ? AND company_id = ?", existingUser.ID, companyID).First(&companyUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists in this company"})
			return
		}

		// Add existing user to company
		companyUser = models.CompanyUser{
			CompanyID: companyID,
			UserID:    existingUser.ID,
			Role:      req.Role,
			IsActive:  true,
			JoinedAt:  time.Now(),
		}

		if err := ah.db.Create(&companyUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to company"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Existing user added to company",
			"user": gin.H{
				"id":       existingUser.ID,
				"username": existingUser.Username,
				"email":    existingUser.Email,
			},
		})
		return
	}

	// Create new user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if req.Role == "" {
		req.Role = "employee"
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
		IsActive: true,
	}

	tx := ah.db.Begin()

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Create company-user relationship
	companyUser := models.CompanyUser{
		CompanyID: companyID,
		UserID:    user.ID,
		Role:      req.Role,
		IsDefault: true,
		IsActive:  true,
		JoinedAt:  time.Now(),
	}

	if err := tx.Create(&companyUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company-user relationship"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func (ah *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	companyID := middleware.GetCompanyID(c)

	var user models.User
	if err := ah.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var employee models.Employee
	ah.db.Preload("Position").Preload("Department").Preload("Currency").
		Where("user_id = ? AND company_id = ?", userID, companyID).First(&employee)

	var companyUser models.CompanyUser
	ah.db.Preload("Company").Where("user_id = ? AND company_id = ?", userID, companyID).First(&companyUser)

	response := gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"last_login": user.LastLogin,
	}

	if companyUser.ID != 0 {
		response["company"] = gin.H{
			"id":   companyUser.Company.ID,
			"name": companyUser.Company.Name,
			"code": companyUser.Company.Code,
			"role": companyUser.Role,
		}
	}

	if employee.ID != 0 {
		response["employee"] = employee
	}

	c.JSON(http.StatusOK, response)
}

func (ah *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := ah.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user.Password = string(hashedPassword)
	if err := ah.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
