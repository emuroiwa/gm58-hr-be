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
	employeeHandler := handlers.NewEmployeeHandler(db, currencyService)
	payrollHandler := handlers.NewPayrollHandler(db, payrollProcessor)
	currencyHandler := handlers.NewCurrencyHandler(db, currencyService)
	positionHandler := handlers.NewPositionHandler(db)
	departmentHandler := handlers.NewDepartmentHandler(db)

	// Public routes
	public := r.Group("/api/v1")
	{
		public.POST("/auth/login", authHandler.Login)
		public.POST("/auth/register", authHandler.Register)

		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now()})
		})
	}

	// Protected routes
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware("jwt-secret"))
	{
		// Auth routes
		auth := protected.Group("/auth")
		{
			auth.GET("/profile", authHandler.GetProfile)
			auth.PUT("/change-password", authHandler.ChangePassword)
		}

		// Employee routes
		employees := protected.Group("/employees")
		{
			employees.GET("", employeeHandler.GetEmployees)
			employees.POST("", employeeHandler.CreateEmployee)
			employees.GET("/:id", employeeHandler.GetEmployee)
			employees.PUT("/:id", employeeHandler.UpdateEmployee)
			employees.DELETE("/:id", employeeHandler.DeleteEmployee)
			employees.GET("/:id/payslips", employeeHandler.GetEmployeePayslips)
		}

		// Department routes
		departments := protected.Group("/departments")
		{
			departments.GET("", departmentHandler.GetDepartments)
			departments.POST("", departmentHandler.CreateDepartment)
			departments.GET("/:id", departmentHandler.GetDepartment)
			departments.PUT("/:id", departmentHandler.UpdateDepartment)
			departments.DELETE("/:id", departmentHandler.DeleteDepartment)
		}

		// Position routes
		positions := protected.Group("/positions")
		{
			positions.GET("", positionHandler.GetPositions)
			positions.POST("", positionHandler.CreatePosition)
			positions.GET("/:id", positionHandler.GetPosition)
			positions.PUT("/:id", positionHandler.UpdatePosition)
			positions.DELETE("/:id", positionHandler.DeletePosition)
			positions.GET("/department/:departmentId", positionHandler.GetPositionsByDepartment)
		}

		// Payroll routes
		payroll := protected.Group("/payroll")
		{
			payroll.POST("/periods", payrollHandler.CreatePeriod)
			payroll.GET("/periods", payrollHandler.GetPeriods)
			payroll.POST("/periods/:periodId/process", payrollHandler.ProcessPayroll)
			payroll.POST("/periods/:periodId/approve", payrollHandler.ApprovePayroll)
			payroll.GET("/periods/:periodId/payslips", payrollHandler.GetPayslips)
			payroll.GET("/periods/:periodId/summary", payrollHandler.GetPayrollSummary)
			payroll.GET("/payslips/:payslipId", payrollHandler.GetPayslip)
		}

		// Currency routes
		currencies := protected.Group("/currencies")
		{
			currencies.GET("", currencyHandler.GetCurrencies)
			currencies.POST("", currencyHandler.CreateCurrency)
			currencies.GET("/exchange-rate", currencyHandler.GetExchangeRate)
			currencies.GET("/convert", currencyHandler.ConvertAmount)
			currencies.POST("/update-rates", currencyHandler.UpdateExchangeRates)
			currencies.GET("/rate-history", currencyHandler.GetExchangeRateHistory)
		}
	}

	// Admin only routes
	admin := protected.Group("/api/v1/admin")
	admin.Use(middleware.RoleMiddleware("admin"))
	{
		// Admin specific routes
	}
}
