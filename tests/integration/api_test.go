package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"gm58-hr-backend/internal/api/routes"
	"gm58-hr-backend/internal/database"
	"gm58-hr-backend/internal/models"
	"gm58-hr-backend/pkg/logger"
	"gm58-hr-backend/pkg/redis"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type APITestSuite struct {
	suite.Suite
	router *gin.Engine
	db     *gorm.DB
	token  string
}

func (suite *APITestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	
	err = database.AutoMigrate(db)
	suite.Require().NoError(err)
	
	suite.seedTestData(db)
	
	suite.router = gin.New()
	suite.db = db
	
	redisClient := &redis.Client{}
	logger := logger.New("info")
	
	routes.SetupRoutes(suite.router, db, redisClient, logger)
	
	suite.token = suite.getAuthToken()
}

func (suite *APITestSuite) seedTestData(db *gorm.DB) {
	usd := models.Currency{
		Code:           "USD",
		Name:           "US Dollar",
		Symbol:         "$",
		IsActive:       true,
		IsBaseCurrency: true,
	}
	db.Create(&usd)
	
	dept := models.Department{
		Name:        "IT",
		Description: "Information Technology",
		IsActive:    true,
	}
	db.Create(&dept)
	
	position := models.Position{
		Title:        "Software Developer",
		DepartmentID: dept.ID,
		MinSalary:    3000,
		MaxSalary:    8000,
		CurrencyID:   usd.ID,
		IsActive:     true,
	}
	db.Create(&position)
	
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
		Role:     "admin",
		IsActive: true,
	}
	db.Create(&user)
}

func (suite *APITestSuite) getAuthToken() string {
	loginData := map[string]interface{}{
		"username": "testuser",
		"password": "password",
	}
	
	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	return response["token"].(string)
}

func (suite *APITestSuite) TestCreateEmployee() {
	employeeData := map[string]interface{}{
		"first_name":    "John",
		"last_name":     "Doe",
		"email":         "john.doe@example.com",
		"phone":         "+1234567890",
		"position_id":   1,
		"department_id": 1,
		"basic_salary":  5000.00,
		"currency_id":   1,
		"hire_date":     "2024-01-01T00:00:00Z",
	}
	
	jsonData, _ := json.Marshal(employeeData)
	req, _ := http.NewRequest("POST", "/api/v1/employees", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)
	
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response models.Employee
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "John", response.FirstName)
	assert.Equal(suite.T(), "Doe", response.LastName)
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
