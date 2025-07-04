package middleware

import (
	"gm58-hr-backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CompanyMiddleware ensures that all requests are scoped to a specific company
func CompanyMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		role := c.GetString("role")

		// Super admins can access any company via query parameter
		if role == "super_admin" {
			companyIDStr := c.Query("company_id")
			if companyIDStr == "" {
				companyIDStr = c.GetHeader("X-Company-ID")
			}
			if companyIDStr != "" {
				companyID, err := strconv.ParseUint(companyIDStr, 10, 32)
				if err == nil {
					c.Set("company_id", uint(companyID))
					c.Set("is_super_admin", true)
					c.Next()
					return
				}
			}
		}

		// For regular users, get company from header or default company
		companyIDStr := c.GetHeader("X-Company-ID")
		if companyIDStr == "" {
			// Get user's default company
			var companyUser models.CompanyUser
			if err := db.Where("user_id = ? AND is_default = ? AND is_active = ?",
				userID, true, true).First(&companyUser).Error; err != nil {
				// If no default, get first active company
				if err := db.Where("user_id = ? AND is_active = ?",
					userID, true).First(&companyUser).Error; err != nil {
					c.JSON(http.StatusForbidden, gin.H{"error": "No active company found for user"})
					c.Abort()
					return
				}
			}
			c.Set("company_id", companyUser.CompanyID)
			c.Set("company_role", companyUser.Role)
		} else {
			// Verify user has access to specified company
			companyID, err := strconv.ParseUint(companyIDStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
				c.Abort()
				return
			}

			var companyUser models.CompanyUser
			if err := db.Where("user_id = ? AND company_id = ? AND is_active = ?",
				userID, uint(companyID), true).First(&companyUser).Error; err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this company"})
				c.Abort()
				return
			}

			c.Set("company_id", uint(companyID))
			c.Set("company_role", companyUser.Role)
		}

		c.Next()
	}
}

// CompanyAdminMiddleware ensures user is an admin for the current company
func CompanyAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		companyRole := c.GetString("company_role")
		isSuperAdmin := c.GetBool("is_super_admin")

		if !isSuperAdmin && companyRole != "company_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Company admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CompanyScope returns a GORM scope that filters by company_id
func CompanyScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		companyID := c.GetUint("company_id")
		if companyID == 0 {
			// This should not happen if middleware is properly set up
			return db.Where("1 = 0") // Return no results
		}
		return db.Where("company_id = ?", companyID)
	}
}

// GetCompanyID is a helper function to get the current company ID from context
func GetCompanyID(c *gin.Context) uint {
	return c.GetUint("company_id")
}

// GetCompanyRole gets the user's role within the current company
func GetCompanyRole(c *gin.Context) string {
	return c.GetString("company_role")
}

// IsSuperAdmin checks if the current user is a super admin
func IsSuperAdmin(c *gin.Context) bool {
	return c.GetBool("is_super_admin")
}
