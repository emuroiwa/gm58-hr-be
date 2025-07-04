package routes

import (
	"gm58-hr-backend/internal/api/handlers"
	"gm58-hr-backend/internal/api/middleware"
	"gm58-hr-backend/internal/services/currency"
	"gm58-hr-backend/internal/services/payroll"
	"gm58-hr-backend/pkg/logger"
	"gm58-hr-backend/pkg/redis"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, redisClient *redis.Client, logger *logger.Logger) {
	// Initialize services
	currencyService := currency.NewCurrencyService(db, "", "")
	payrollProcessor := payroll.NewPayrollProcessor(db, currencyService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, "jwt-secret")
	companyHandler := handlers.NewCompanyHandler(db)
	employeeHandler := handlers.NewEmployeeHandler(db, currencyService)
	payrollHandler := handlers.NewPayrollHandler(db, payrollProcessor)
	currencyHandler := handlers.NewCurrencyHandler(db, currencyService)
	positionHandler := handlers.NewPositionHandler(db)
	departmentHandler := handlers.NewDepartmentHandler(db)

	// Public routes (no authentication required)
	public := r.Group("/api/v1")
	{
		// Auth endpoints
		public.POST("/auth/login", authHandler.Login)
		public.POST("/auth/register-company", authHandler.RegisterCompany) // New company registration

		// Health check
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now()})
		})
	}

	// Authenticated routes (require JWT token)
	auth := r.Group("/api/v1")
	auth.Use(middleware.AuthMiddleware("jwt-secret"))
	{
		// Auth routes that don't require company context
		authRoutes := auth.Group("/auth")
		{
			authRoutes.GET("/profile", authHandler.GetProfile)
			authRoutes.PUT("/change-password", authHandler.ChangePassword)
			authRoutes.GET("/companies", authHandler.GetUserCompanies)    // List user's companies
			authRoutes.POST("/switch-company", authHandler.SwitchCompany) // Switch active company
		}
	}

	// Company-scoped routes (require company context)
	company := r.Group("/api/v1")
	company.Use(middleware.AuthMiddleware("jwt-secret"))
	company.Use(middleware.CompanyMiddleware(db))
	{
		// Company management
		companyRoutes := company.Group("/company")
		{
			companyRoutes.GET("", companyHandler.GetCompany)
			companyRoutes.PUT("", middleware.CompanyAdminMiddleware(), companyHandler.UpdateCompany)
			companyRoutes.GET("/settings", companyHandler.GetCompanySettings)
			companyRoutes.PUT("/settings", middleware.CompanyAdminMiddleware(), companyHandler.UpdateCompanySettings)
			companyRoutes.GET("/stats", companyHandler.GetCompanyStats)

			// Company user management (admin only)
			companyRoutes.GET("/users", middleware.CompanyAdminMiddleware(), companyHandler.GetCompanyUsers)
			companyRoutes.POST("/users", middleware.CompanyAdminMiddleware(), companyHandler.AddUserToCompany)
			companyRoutes.PUT("/users/:userId/role", middleware.CompanyAdminMiddleware(), companyHandler.UpdateUserRole)
			companyRoutes.DELETE("/users/:userId", middleware.CompanyAdminMiddleware(), companyHandler.RemoveUserFromCompany)
		}

		// User registration within company context
		company.POST("/auth/register", authHandler.Register)

		// Employee routes
		employees := company.Group("/employees")
		{
			employees.GET("", employeeHandler.GetEmployees)
			employees.POST("", employeeHandler.CreateEmployee)
			employees.GET("/:id", employeeHandler.GetEmployee)
			employees.PUT("/:id", employeeHandler.UpdateEmployee)
			employees.DELETE("/:id", employeeHandler.DeleteEmployee)
			employees.GET("/:id/payslips", employeeHandler.GetEmployeePayslips)
		}

		// Department routes
		departments := company.Group("/departments")
		{
			departments.GET("", departmentHandler.GetDepartments)
			departments.POST("", departmentHandler.CreateDepartment)
			departments.GET("/:id", departmentHandler.GetDepartment)
			departments.PUT("/:id", departmentHandler.UpdateDepartment)
			departments.DELETE("/:id", departmentHandler.DeleteDepartment)
		}

		// Position routes
		positions := company.Group("/positions")
		{
			positions.GET("", positionHandler.GetPositions)
			positions.POST("", positionHandler.CreatePosition)
			positions.GET("/:id", positionHandler.GetPosition)
			positions.PUT("/:id", positionHandler.UpdatePosition)
			positions.DELETE("/:id", positionHandler.DeletePosition)
			positions.GET("/department/:departmentId", positionHandler.GetPositionsByDepartment)
		}

		// Payroll routes
		payroll := company.Group("/payroll")
		{
			payroll.POST("/periods", payrollHandler.CreatePeriod)
			payroll.GET("/periods", payrollHandler.GetPeriods)
			payroll.POST("/periods/:periodId/process", payrollHandler.ProcessPayroll)
			payroll.POST("/periods/:periodId/approve", payrollHandler.ApprovePayroll)
			payroll.GET("/periods/:periodId/payslips", payrollHandler.GetPayslips)
			payroll.GET("/periods/:periodId/summary", payrollHandler.GetPayrollSummary)
			payroll.GET("/payslips/:payslipId", payrollHandler.GetPayslip)
		}

		// Currency routes (some are global, some are company-specific)
		currencies := company.Group("/currencies")
		{
			currencies.GET("", currencyHandler.GetCurrencies)
			currencies.POST("", currencyHandler.CreateCurrency)
			currencies.GET("/exchange-rate", currencyHandler.GetExchangeRate)
			currencies.GET("/convert", currencyHandler.ConvertAmount)
			currencies.POST("/update-rates", currencyHandler.UpdateExchangeRates)
			currencies.GET("/rate-history", currencyHandler.GetExchangeRateHistory)
		}
	}

	// Super admin routes
	superAdmin := r.Group("/api/v1/admin")
	superAdmin.Use(middleware.AuthMiddleware("jwt-secret"))
	superAdmin.Use(middleware.RoleMiddleware("super_admin"))
	{
		// Company management for super admins
		superAdmin.GET("/companies", companyHandler.ListCompanies)
		superAdmin.GET("/companies/:id", func(c *gin.Context) {
			// Allow super admin to view any company
			c.Set("company_id", c.Param("id"))
			companyHandler.GetCompany(c)
		})

		// Global currency management
		superAdmin.POST("/currencies", currencyHandler.CreateCurrency)
		superAdmin.POST("/currencies/update-all-rates", currencyHandler.UpdateExchangeRates)
	}
}
